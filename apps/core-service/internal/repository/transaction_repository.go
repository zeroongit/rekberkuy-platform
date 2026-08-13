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

// CreateTransaction stores the pure-IDR escrow transaction master data into the Supabase database
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
		return fmt.Errorf("failed to create escrow transaction data: %w", err)
	}
	return nil
}

// GetTransactionByID fetches the transaction detail data with row locking (FOR UPDATE)
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
			return nil, fmt.Errorf("transaction with id %s not found", id)
		}
		return nil, err
	}

	return &tx, nil
}

// GetTransactionByMidtransOrderID fetches a transaction by its Midtrans order id (webhook).
func (r *TransactionRepository) GetTransactionByMidtransOrderID(ctx context.Context, orderID string) (*domain.Transaction, error) {
	query := `
		SELECT id, buyer_id, seller_id, type, status, amount_base, shipping_fee,
		       service_fee, midtrans_fee, amount_gross, amount_net,
		       midtrans_order_id, idempotency_key, payment_method,
		       blockchain_tx_hash, blockchain_logged_at, created_at, updated_at
		FROM transactions
		WHERE midtrans_order_id = $1
	`
	// Lock the row inside a transaction so concurrent webhook retries cannot both
	// read WAITING_PAYMENT and double-debit the buyer (anti double-spending).
	if r.tx != nil {
		query += " FOR UPDATE"
	}
	var tx domain.Transaction
	scanRow := func(row interface{ Scan(...interface{}) error }) error {
		return row.Scan(
			&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.Type, &tx.Status, &tx.AmountBase, &tx.ShippingFee,
			&tx.ServiceFee, &tx.MidtransFee, &tx.AmountGross, &tx.AmountNet,
			&tx.MidtransOrderID, &tx.IdempotencyKey, &tx.PaymentMethod,
			&tx.BlockchainTxHash, &tx.BlockchainLoggedAt, &tx.CreatedAt, &tx.UpdatedAt,
		)
	}
	var err error
	if r.tx != nil {
		err = scanRow(r.tx.QueryRowContext(ctx, query, orderID))
	} else {
		err = scanRow(r.db.QueryRowContext(ctx, query, orderID))
	}
	if err != nil {
		return nil, fmt.Errorf("transaction with midtrans order id %s not found: %w", orderID, err)
	}
	return &tx, nil
}

// UpdateTransactionStatus processes the State Machine transition (Universal State Machine)
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
		return fmt.Errorf("failed to update transaction status: %w", err)
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
	var rows *sql.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query)
	} else {
		rows, err = r.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query expired transactions: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan expired transaction ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *TransactionRepository) GetMilestoneByID(ctx context.Context, id string) (*domain.ServiceMilestone, error) {
	query := `SELECT id, transaction_id, milestone_index, title, amount, status FROM service_milestones WHERE id = $1`
	var m domain.ServiceMilestone
	var err error
	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.TransactionID, &m.MilestoneIndex, &m.Title, &m.Amount, &m.Status)
	} else {
		err = r.db.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.TransactionID, &m.MilestoneIndex, &m.Title, &m.Amount, &m.Status)
	}
	if err != nil {
		return nil, fmt.Errorf("milestone not found: %w", err)
	}
	return &m, nil
}

func (r *TransactionRepository) UpdateMilestoneStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE service_milestones SET status = $1, released_at = NOW() WHERE id = $2`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, status, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, id)
	}
	return err
}

func (r *TransactionRepository) GetEventVendorPayoutsByTxID(ctx context.Context, txID string) ([]domain.EventVendorPayout, error) {
	query := `SELECT id, transaction_id, vendor_user_id, vendor_name, amount_requested, status FROM event_vendor_payouts WHERE transaction_id = $1`
	var rows *sql.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query, txID)
	} else {
		rows, err = r.db.QueryContext(ctx, query, txID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []domain.EventVendorPayout
	for rows.Next() {
		var p domain.EventVendorPayout
		if err := rows.Scan(&p.ID, &p.TransactionID, &p.VendorUserID, &p.VendorName, &p.AmountRequested, &p.Status); err != nil {
			return nil, err
		}
		payouts = append(payouts, p)
	}
	return payouts, nil
}

func (r *TransactionRepository) UpdateEventVendorPayoutStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE event_vendor_payouts SET status = $1, reviewed_at = NOW() WHERE id = $2`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, status, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, id)
	}
	return err
}

// GetEventVendorPayoutByID fetches a single event vendor payout by its id.
func (r *TransactionRepository) GetEventVendorPayoutByID(ctx context.Context, payoutID string) (*domain.EventVendorPayout, error) {
	query := `SELECT id, transaction_id, vendor_user_id, vendor_name, amount_requested, status FROM event_vendor_payouts WHERE id = $1`
	var p domain.EventVendorPayout
	var err error
	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, payoutID).Scan(&p.ID, &p.TransactionID, &p.VendorUserID, &p.VendorName, &p.AmountRequested, &p.Status)
	} else {
		err = r.db.QueryRowContext(ctx, query, payoutID).Scan(&p.ID, &p.TransactionID, &p.VendorUserID, &p.VendorName, &p.AmountRequested, &p.Status)
	}
	if err != nil {
		return nil, fmt.Errorf("event vendor payout %s not found: %w", payoutID, err)
	}
	return &p, nil
}

// MarkEventVendorPayoutDisbursed records completion of an external vendor's bank
// payout. The transfer itself happens out-of-band; this only updates the record
// (see ADR-0002 — admin-triggered disbursement, no in-system money movement).
func (r *TransactionRepository) MarkEventVendorPayoutDisbursed(ctx context.Context, payoutID string, adminID string) error {
	query := `UPDATE event_vendor_payouts SET status = $1, reviewed_by = $2, reviewed_at = NOW(), disbursed_at = NOW() WHERE id = $3`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, domain.VendorPayoutDisbursed, adminID, payoutID)
	} else {
		_, err = r.db.ExecContext(ctx, query, domain.VendorPayoutDisbursed, adminID, payoutID)
	}
	if err != nil {
		return fmt.Errorf("failed to mark vendor payout disbursed: %w", err)
	}
	return nil
}

// UpdateBlockchainLog stores the on-chain audit-log hash & recording timestamp.
func (r *TransactionRepository) UpdateBlockchainLog(ctx context.Context, txID string, txHash string) error {
	query := `
		UPDATE transactions
		SET blockchain_tx_hash = $1, blockchain_logged_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, txHash, txID)
	} else {
		_, err = r.db.ExecContext(ctx, query, txHash, txID)
	}
	if err != nil {
		return fmt.Errorf("failed to save blockchain tx hash: %w", err)
	}
	return nil
}
