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
	tx *sql.Tx
}

// NewWalletRepository menginisialisasi adapter database untuk RekberPay Wallet
func NewWalletRepository(db *sql.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

// ExecuteInTransaction menjalankan blok fungsi di dalam satu database transaction (ACID)
func (r *walletRepository) ExecuteInTransaction(ctx context.Context, fn func(domain.WalletRepository) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi database: %w", err)
	}

	txRepo := &walletRepository{db: r.db, tx: tx}

	if err := fn(txRepo); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetBalance mengambil data saldo dan status kebekuan dompet user dari Supabase
func (r *walletRepository) GetBalance(ctx context.Context, userID string) (*domain.RekberPayWallet, error) {
	query := `
		SELECT user_id, balance, is_frozen, updated_at 
		FROM rekberpay_wallets 
		WHERE user_id = $1
	`

	var wallet domain.RekberPayWallet
	var err error

	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, userID).Scan(
			&wallet.UserID, &wallet.Balance, &wallet.IsFrozen, &wallet.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, userID).Scan(
			&wallet.UserID, &wallet.Balance, &wallet.IsFrozen, &wallet.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet tidak ditemukan untuk user id: %s", userID)
		}
		return nil, err
	}

	return &wallet, nil
}

// CreateWallet membuat dompet RekberPay baru saat user selesai registrasi
func (r *walletRepository) CreateWallet(ctx context.Context, userID string) error {
	query := `
		INSERT INTO rekberpay_wallets (user_id, balance, is_frozen, updated_at)
		VALUES ($1, 0, false, NOW())
		ON CONFLICT (user_id) DO NOTHING
	`

	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, userID)
	} else {
		_, err = r.db.ExecContext(ctx, query, userID)
	}

	if err != nil {
		return fmt.Errorf("gagal membuat wallet baru: %w", err)
	}

	return nil
}

// UpdateBalanceTx mengeksekusi mutasi saldo secara aman dengan perlindungan Race Condition
func (r *walletRepository) UpdateBalanceTx(ctx context.Context, txRecord *domain.RekberPayTransaction, amountModifier int64) error {
	var tx *sql.Tx
	var err error
	isNestedTx := r.tx != nil

	if isNestedTx {
		tx = r.tx
	} else {
		tx, err = r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return fmt.Errorf("gagal memulai transaksi database: %w", err)
		}
	}

	defer func() {
		if err != nil && !isNestedTx {
			_ = tx.Rollback()
		}
	}()

	// 1. Kunci baris wallet user (SELECT FOR UPDATE)
	var currentBalance int64
	var isFrozen bool
	lockQuery := `
		SELECT balance, is_frozen FROM rekberpay_wallets WHERE user_id = $1 FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, lockQuery, txRecord.WalletID).Scan(&currentBalance, &isFrozen)
	if err != nil {
		return fmt.Errorf("gagal mengunci data wallet untuk mutasi: %w", err)
	}

	if isFrozen {
		return fmt.Errorf("transaksi ditolak: wallet user %s sedang dibekukan", txRecord.WalletID)
	}

	// 2. Kalkulasi jatah potongan saldo beserta biaya withdraw flat Rp7.500
	totalDeduction := amountModifier
	if amountModifier < 0 {
		if txRecord.Type == domain.TxWithdraw {
			totalDeduction = amountModifier - txRecord.AdminFee
		}
		if (currentBalance + totalDeduction) < 0 {
			return errors.New("transaksi ditolak: saldo RekberPay tidak mencukupi untuk nominal transaksi beserta biaya admin")
		}
	}

	// 3. Update tabel saldo master
	updateWalletQuery := `
		UPDATE rekberpay_wallets SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2
	`
	_, err = tx.ExecContext(ctx, updateWalletQuery, totalDeduction, txRecord.WalletID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui saldo wallet: %w", err)
	}

	// 4. Catat riwayat mutasi finansial ke tabel rekberpay_transactions
	insertLogQuery := `
		INSERT INTO rekberpay_transactions (
			id, wallet_id, type, status, amount, admin_fee, 
			platform_net_profit, reference_transaction_id, 
			midtrans_topup_id, description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`
	_, err = tx.ExecContext(ctx, insertLogQuery,
		txRecord.ID, txRecord.WalletID, txRecord.Type, txRecord.Status,
		txRecord.Amount, txRecord.AdminFee, txRecord.PlatformNetProfit,
		txRecord.ReferenceTransactionID, txRecord.MidtransTopUpID, txRecord.Description,
	)
	if err != nil {
		return fmt.Errorf("gagal mencatat histori mutasi transaksi: %w", err)
	}

	if !isNestedTx {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("gagal melakukan commit transaksi keuangan: %w", err)
		}
	}

	return nil
}

// GetAllUsersForCRMEvaluation mengambil seluruh data profil pengguna (Patuhi kontrak kembalian []*domain.UserProfile)
func (r *walletRepository) GetAllUsersForCRMEvaluation(ctx context.Context) ([]*domain.UserProfile, error) {
	query := `
		SELECT id, username, full_name, role, phone_number, created_at, updated_at FROM user_profiles
	`
	var rows *sql.Rows
	var err error

	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query)
	} else {
		rows, err = r.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data profil evaluasi CRM: %w", err)
	}
	defer rows.Close()

	var users []*domain.UserProfile
	for rows.Next() {
		var user domain.UserProfile
		err := rows.Scan(
			&user.ID, &user.Username, &user.FullName, &user.Role, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan data user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetCRMLoyaltyByUserID mengambil jatah profile loyalitas kasta user
func (r *walletRepository) GetCRMLoyaltyByUserID(ctx context.Context, userID string) (*domain.CRMLoyalty, error) {
	query := `
		SELECT user_id, total_points, current_tier, total_spent_fiat, rolling_3_month_gmv, 
		       current_month_gmv, max_item_price_sold, total_completed_services, 
		       total_completed_events, consecutive_failed_months, tier_evaluation_started_at, 
		       last_month_evaluated_at, updated_at
		FROM crm_loyalty WHERE user_id = $1
	`
	var crm domain.CRMLoyalty
	var err error

	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, userID).Scan(
			&crm.UserID, &crm.TotalPoints, &crm.CurrentTier, &crm.TotalSpentFiat, &crm.Rolling3MonthGMV,
			&crm.CurrentMonthGmv, &crm.MaxItemPriceSold, &crm.TotalCompletedServices,
			&crm.TotalCompletedEvents, &crm.ConsecutiveFailedMonths, &crm.TierEvaluationStartedAt,
			&crm.LastMonthEvaluatedAt, &crm.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, userID).Scan(
			&crm.UserID, &crm.TotalPoints, &crm.CurrentTier, &crm.TotalSpentFiat, &crm.Rolling3MonthGMV,
			&crm.CurrentMonthGmv, &crm.MaxItemPriceSold, &crm.TotalCompletedServices,
			&crm.TotalCompletedEvents, &crm.ConsecutiveFailedMonths, &crm.TierEvaluationStartedAt,
			&crm.LastMonthEvaluatedAt, &crm.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("crm profile tidak ditemukan untuk user id: %s", userID)
		}
		return nil, err
	}

	return &crm, nil
}

// UpdateCRMLoyalty menyimpan pembaruan kasta terbaru hasil evaluasi bulanan
func (r *walletRepository) UpdateCRMLoyalty(ctx context.Context, crmProfile *domain.CRMLoyalty) error {
	query := `
		UPDATE crm_loyalty 
		SET current_tier = $1, consecutive_failed_months = $2, last_month_evaluated_at = NOW(), updated_at = NOW()
		WHERE user_id = $3
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, crmProfile.CurrentTier, crmProfile.ConsecutiveFailedMonths, crmProfile.UserID)
	} else {
		_, err = r.db.ExecContext(ctx, query, crmProfile.CurrentTier, crmProfile.ConsecutiveFailedMonths, crmProfile.UserID)
	}

	if err != nil {
		return fmt.Errorf("gagal memperbarui data crm kasta user: %w", err)
	}

	return nil
}