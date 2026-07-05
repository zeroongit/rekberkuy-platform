package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type TransactionRepository struct {
	db *sql.DB
	tx *sql.Tx
}


func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction menyimpan data master transaksi escrow murni IDR ke database Supabase
func (r *TransactionRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, buyer_id, seller_id, type, status, amount_base, shipping_fee, 
			service_fee, midtrans_fee, amount_gross, amount_net, 
			midtrans_order_id, idempotency_key, payment_method, 
			blockchain_tx_hash, blockchain_logged_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
	`

	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query,
			tx.ID, tx.BuyerID, tx.SellerID, tx.Type, tx.Status, tx.AmountBase, tx.ShippingFee,
			tx.ServiceFee, tx.MidtransFee, tx.AmountGross, tx.AmountNet,
			tx.MidtransOrderID, tx.IdempotencyKey, tx.PaymentMethod,
			tx.BlockchainTxHash, tx.BlockchainLoggedAt,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			tx.ID, tx.BuyerID, tx.SellerID, tx.Type, tx.Status, tx.AmountBase, tx.ShippingFee,
			tx.ServiceFee, tx.MidtransFee, tx.AmountGross, tx.AmountNet,
			tx.MidtransOrderID, tx.IdempotencyKey, tx.PaymentMethod,
			tx.BlockchainTxHash, tx.BlockchainLoggedAt,
		)
	}

	if err != nil {
		return fmt.Errorf("gagal membuat data transaksi escrow: %w", err)
	}
	return nil
}

// GetTransactionByID mengambil detail data transaksi lengkap dengan pengunci baris (FOR UPDATE)
func (r *TransactionRepository) GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `
		SELECT id, buyer_id, seller_id, type, status, amount_base, shipping_fee, 
		       service_fee, midtrans_fee, amount_gross, amount_net, 
		       midtrans_order_id, idempotency_key, payment_method, 
		       blockchain_tx_hash, blockchain_logged_at, created_at, updated_at
		FROM transactions 
		WHERE id = $1
	`
	
	if r.tx != nil {
		query += " FOR UPDATE"
	}

	var tx domain.Transaction
	var err error

	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, id).Scan(
			&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.Type, &tx.Status, &tx.AmountBase, &tx.ShippingFee,
			&tx.ServiceFee, &tx.MidtransFee, &tx.AmountGross, &tx.AmountNet,
			&tx.MidtransOrderID, &tx.IdempotencyKey, &tx.PaymentMethod,
			&tx.BlockchainTxHash, &tx.BlockchainLoggedAt, &tx.CreatedAt, &tx.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, id).Scan(
			&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.Type, &tx.Status, &tx.AmountBase, &tx.ShippingFee,
			&tx.ServiceFee, &tx.MidtransFee, &tx.AmountGross, &tx.AmountNet,
			&tx.MidtransOrderID, &tx.IdempotencyKey, &tx.PaymentMethod,
			&tx.BlockchainTxHash, &tx.BlockchainLoggedAt, &tx.CreatedAt, &tx.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaksi dengan id %s tidak ditemukan", id)
		}
		return nil, err
	}

	return &tx, nil
}

// UpdateTransactionStatus memproses transisi State Machine (Universal State Machine)
func (r *TransactionRepository) UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	query := `
		UPDATE transactions 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2
	`

	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, status, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, id)
	}

	if err != nil {
		return fmt.Errorf("gagal memperbarui status transaksi: %w", err)
	}
	return nil
}


func (r *TransactionRepository) GetExpiredLockedTransactions(ctx context.Context) ([]string, error) {
	query := `
		SELECT t.id 
		FROM transactions t
		JOIN transaction_goods tg ON t.id = tg.transaction_id
		WHERE t.status = 'FUNDS_LOCKED' AND tg.auto_confirm_deadline <= NOW()
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal kueri transaksi expired: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("gagal scan ID transaksi expired: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *TransactionRepository) GetMilestoneByID(ctx context.Context, id string) (*domain.ServiceMilestone, error) {
	query := `SELECT id, transaction_id, milestone_index, title, amount, status FROM service_milestones WHERE id = $1`
	var m domain.ServiceMilestone
	err := r.db.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.TransactionID, &m.MilestoneIndex, &m.Title, &m.Amount, &m.Status)
	if err != nil {
		return nil, fmt.Errorf("milestone tidak ditemukan: %w", err)
	}
	return &m, nil
}

func (r *TransactionRepository) UpdateMilestoneStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE service_milestones SET status = $1, released_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *TransactionRepository) GetEventVendorPayoutsByTxID(ctx context.Context, txID string) ([]domain.EventVendorPayout, error) {
	query := `SELECT id, transaction_id, vendor_name, amount_requested, status FROM event_vendor_payouts WHERE transaction_id = $1`
	rows, err := r.db.QueryContext(ctx, query, txID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []domain.EventVendorPayout
	for rows.Next() {
		var p domain.EventVendorPayout
		if err := rows.Scan(&p.ID, &p.TransactionID, &p.VendorName, &p.AmountRequested, &p.Status); err != nil {
			return nil, err
		}
		payouts = append(payouts, p)
	}
	return payouts, nil
}

func (r *TransactionRepository) UpdateEventVendorPayoutStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE event_vendor_payouts SET status = $1, reviewed_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}