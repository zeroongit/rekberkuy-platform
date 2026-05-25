package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"


	"rekberkuy/core-service/internal/domain"
)

type walletRepository struct {
	db *sql.DB
}

// NewWalletRepository menginisialisasi adapter database untuk RekberPay Wallet
func NewWalletRepository(db *sql.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

// GetBalance mengambil data saldo dan status kebekuan dompet user dari Supabase
func (r *walletRepository) GetBalance(ctx context.Context, userID string) (*domain.RekberPayWallet, error) {
	query := `
		SELECT user_id, balance, is_frozen, updated_at 
		FROM rekberpay_wallets 
		WHERE user_id = $1
	`
	
	var wallet domain.RekberPayWallet
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&wallet.UserID,
		&wallet.Balance,
		&wallet.IsFrozen,
		&wallet.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet tidak ditemukan untuk user id: %s", userID)
		}
		return nil, err
	}
	
	return &wallet, nil
}

// CreateWallet membuat dompet RekberPay baru (biasanya dipicu otomatis saat user selesai registrasi)
func (r *walletRepository) CreateWallet(ctx context.Context, userID string) error {
	query := `
		INSERT INTO rekberpay_wallets (user_id, balance, is_frozen, updated_at)
		VALUES ($1, 0, false, NOW())
		ON CONFLICT (user_id) DO NOTHING
	`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("gagal membuat wallet baru: %w", err)
	}
	
	return nil
}

// UpdateBalanceTx mengeksekusi mutasi saldo secara ACID menggunakan database TRANSACTION.
// amountModifier bisa bernilai positif (top-up/receive) atau negatif (payment/withdraw).
func (r *walletRepository) UpdateBalanceTx(ctx context.Context, txRecord *domain.RekberPayTransaction, amountModifier int64) error {
	// 1. Memulai database transaction (Tx) untuk menjamin atomisitas data finansial
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi database: %w", err)
	}
	
	// Pastikan melakukan rollback jika terjadi panic atau error di tengah jalan.
	// Menggunakan blank identifier _ untuk membungkam warning linter unhandled error.
	defer func() {
		_ = tx.Rollback()
	}()

	// 2. Kunci baris wallet user yang bersangkutan (SELECT FOR UPDATE) untuk mencegah Race Condition / Double Spending
	var currentBalance int64
	var isFrozen bool
	lockQuery := `
		SELECT balance, is_frozen 
		FROM rekberpay_wallets 
		WHERE user_id = $1 
		FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, lockQuery, txRecord.WalletID).Scan(&currentBalance, &isFrozen)
	if err != nil {
		return fmt.Errorf("gagal mengunci data wallet untuk mutasi: %w", err)
	}

	// 3. Validasi keamanan: Jika dompet dibekukan oleh sistem, batalkan semua mutasi keluar/masuk
	if isFrozen {
		return fmt.Errorf("transaksi ditolak: wallet user %s sedang dibekukan karena sengketa/sanksi", txRecord.WalletID)
	}

	// 4. Validasi kecukupan saldo jika melakukan penarikan/pembayaran (modifier negatif)
	if amountModifier < 0 && (currentBalance + amountModifier) < 0 {
		return errors.New("transaksi ditolak: saldo RekberPay tidak mencukupi")
	}

	// 5. Eksekusi pembaruan saldo di tabel master wallet
	updateWalletQuery := `
		UPDATE rekberpay_wallets 
		SET balance = balance + $1, updated_at = NOW() 
		WHERE user_id = $2
	`
	_, err = tx.ExecContext(ctx, updateWalletQuery, amountModifier, txRecord.WalletID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui saldo wallet: %w", err)
	}

	// 6. Catat histori mutasi ke tabel rekberpay_transactions sebagai audit log finansial
	insertLogQuery := `
		INSERT INTO rekberpay_transactions (
			id, wallet_id, type, status, amount, admin_fee, 
			platform_net_profit, reference_transaction_id, 
			midtrans_topup_id, description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`
	_, err = tx.ExecContext(ctx, insertLogQuery,
		txRecord.ID,
		txRecord.WalletID,
		txRecord.Type,
		txRecord.Status,
		txRecord.Amount,
		txRecord.AdminFee,
		txRecord.PlatformNetProfit,
		txRecord.ReferenceTransactionID,
		txRecord.MidtransTopUpID,
		txRecord.Description,
	)
	if err != nil {
		return fmt.Errorf("gagal mencatat histori mutasi transaksi: %w", err)
	}

	// 7. Jika semua langkah di atas aman tanpa kendala, commit transaksi ke Supabase
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal melakukan commit transaksi keuangan: %w", err)
	}

	return nil
}