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

// CreateGoodsDetail stores the goods-specific detail container row. Must run
// inside the same UnitOfWork as CreateTransaction for the master row.
func (r *TransactionRepository) CreateGoodsDetail(ctx context.Context, tg *domain.TransactionGoods) error {
	query := `
		INSERT INTO transaction_goods (
			transaction_id, sub_sub_category_id, shipping_courier, shipping_tracking_number,
			shipping_address, auto_confirm_deadline
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query,
			tg.TransactionID, tg.SubSubCategoryID, tg.ShippingCourier, tg.ShippingTrackingNumber,
			tg.ShippingAddress, tg.AutoConfirmDeadline,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			tg.TransactionID, tg.SubSubCategoryID, tg.ShippingCourier, tg.ShippingTrackingNumber,
			tg.ShippingAddress, tg.AutoConfirmDeadline,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to create goods transaction detail: %w", err)
	}
	return nil
}

// CreateServicesDetail stores the services detail container plus its milestone
// breakdown in one call so they commit atomically with the master row.
func (r *TransactionRepository) CreateServicesDetail(ctx context.Context, ts *domain.TransactionServices, milestones []domain.ServiceMilestone) error {
	query := `
		INSERT INTO transaction_services (
			transaction_id, sub_sub_category_id, project_deadline, brief_description
		) VALUES ($1, $2, $3, $4)
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query,
			ts.TransactionID, ts.SubSubCategoryID, ts.ProjectDeadline, ts.BriefDescription,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			ts.TransactionID, ts.SubSubCategoryID, ts.ProjectDeadline, ts.BriefDescription,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to create services transaction detail: %w", err)
	}

	for _, m := range milestones {
		mQuery := `
			INSERT INTO service_milestones (id, transaction_id, milestone_index, title, amount, status, created_at)
			VALUES ($1, $2, $3, $4, $5, 'PENDING', NOW())
		`
		if r.tx != nil {
			_, err = r.tx.ExecContext(ctx, mQuery, m.ID, m.TransactionID, m.MilestoneIndex, m.Title, m.Amount)
		} else {
			_, err = r.db.ExecContext(ctx, mQuery, m.ID, m.TransactionID, m.MilestoneIndex, m.Title, m.Amount)
		}
		if err != nil {
			return fmt.Errorf("failed to create service milestone %d: %w", m.MilestoneIndex, err)
		}
	}
	return nil
}

// CreateEventsDetail stores the event detail container plus the pledged vendor
// allocations in one call so they commit atomically with the master row.
func (r *TransactionRepository) CreateEventsDetail(ctx context.Context, te *domain.TransactionEvents, allocations []domain.EventVendorAllocation) error {
	query := `
		INSERT INTO transaction_events (
			transaction_id, sub_sub_category_id, event_name, event_start_time, event_end_time, ticket_quantity_total
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query,
			te.TransactionID, te.SubSubCategoryID, te.EventName, te.EventStartTime, te.EventEndTime, te.TicketQuantityTotal,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			te.TransactionID, te.SubSubCategoryID, te.EventName, te.EventStartTime, te.EventEndTime, te.TicketQuantityTotal,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to create event transaction detail: %w", err)
	}

	for _, a := range allocations {
		aQuery := `
			INSERT INTO event_vendor_allocations (id, transaction_id, vendor_id, allocated_amount, actual_paid_amount, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`
		if r.tx != nil {
			_, err = r.tx.ExecContext(ctx, aQuery, a.ID, a.TransactionID, a.VendorID, a.AllocatedAmount, a.ActualPaidAmount, a.Status)
		} else {
			_, err = r.db.ExecContext(ctx, aQuery, a.ID, a.TransactionID, a.VendorID, a.AllocatedAmount, a.ActualPaidAmount, a.Status)
		}
		if err != nil {
			return fmt.Errorf("failed to create vendor allocation for %s: %w", a.VendorID, err)
		}
	}
	return nil
}

// GetActiveEventVendorPayoutsTotal sums amount_requested over payouts that have
// not reached the terminal DISBURSED state — the amount still claiming escrow.
func (r *TransactionRepository) GetActiveEventVendorPayoutsTotal(ctx context.Context, txID string) (int64, error) {
	const query = `SELECT COALESCE(SUM(amount_requested), 0) FROM event_vendor_payouts WHERE transaction_id = $1 AND status <> $2`
	var total int64
	var err error
	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, txID, domain.VendorPayoutDisbursed).Scan(&total)
	} else {
		err = r.db.QueryRowContext(ctx, query, txID, domain.VendorPayoutDisbursed).Scan(&total)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to sum active vendor payouts for %s: %w", txID, err)
	}
	return total, nil
}

// GetReleasedMilestonesTotalByTxID sums the amounts of milestones already
// RELEASED on a services transaction (0 for goods/events). Used to compute
// remaining escrow during a disputed refund.
func (r *TransactionRepository) GetReleasedMilestonesTotalByTxID(ctx context.Context, transactionID string) (int64, error) {
	const query = `SELECT COALESCE(SUM(amount), 0) FROM service_milestones WHERE transaction_id = $1 AND status = 'RELEASED'`
	var total int64
	var err error
	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, transactionID).Scan(&total)
	} else {
		err = r.db.QueryRowContext(ctx, query, transactionID).Scan(&total)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to sum released milestones for %s: %w", transactionID, err)
	}
	return total, nil
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

// ListTransactionsByUser returns the transactions where the user is buyer or
// seller, newest first (paginated).
func (r *TransactionRepository) ListTransactionsByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Transaction, error) {
	query := `
		SELECT id, buyer_id, seller_id, type, status, amount_base, shipping_fee,
		       service_fee, midtrans_fee, amount_gross, amount_net,
		       midtrans_order_id, idempotency_key, payment_method,
		       blockchain_tx_hash, blockchain_logged_at, created_at, updated_at
		FROM transactions
		WHERE buyer_id = $1 OR seller_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	var rows *sql.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query, userID, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, query, userID, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions for user: %w", err)
	}
	defer rows.Close()

	out := []domain.Transaction{}
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.Type, &tx.Status, &tx.AmountBase, &tx.ShippingFee,
			&tx.ServiceFee, &tx.MidtransFee, &tx.AmountGross, &tx.AmountNet,
			&tx.MidtransOrderID, &tx.IdempotencyKey, &tx.PaymentMethod,
			&tx.BlockchainTxHash, &tx.BlockchainLoggedAt, &tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		out = append(out, tx)
	}
	return out, rows.Err()
}

// GetGoodsDetailByTxID fetches the goods detail container of a transaction.
func (r *TransactionRepository) GetGoodsDetailByTxID(ctx context.Context, txID string) (*domain.TransactionGoods, error) {
	query := `
		SELECT transaction_id, sub_sub_category_id, shipping_courier, shipping_tracking_number,
		       shipping_address, auto_confirm_deadline
		FROM transaction_goods WHERE transaction_id = $1
	`
	var tg domain.TransactionGoods
	row := func() interface{ Scan(...interface{}) error } {
		if r.tx != nil {
			return r.tx.QueryRowContext(ctx, query, txID)
		}
		return r.db.QueryRowContext(ctx, query, txID)
	}
	if err := row().Scan(
		&tg.TransactionID, &tg.SubSubCategoryID, &tg.ShippingCourier, &tg.ShippingTrackingNumber,
		&tg.ShippingAddress, &tg.AutoConfirmDeadline,
	); err != nil {
		return nil, fmt.Errorf("goods detail for %s not found: %w", txID, err)
	}
	return &tg, nil
}

// GetServicesDetailByTxID fetches the services detail container of a transaction.
func (r *TransactionRepository) GetServicesDetailByTxID(ctx context.Context, txID string) (*domain.TransactionServices, error) {
	query := `
		SELECT transaction_id, sub_sub_category_id, project_deadline, brief_description
		FROM transaction_services WHERE transaction_id = $1
	`
	var ts domain.TransactionServices
	row := func() interface{ Scan(...interface{}) error } {
		if r.tx != nil {
			return r.tx.QueryRowContext(ctx, query, txID)
		}
		return r.db.QueryRowContext(ctx, query, txID)
	}
	if err := row().Scan(&ts.TransactionID, &ts.SubSubCategoryID, &ts.ProjectDeadline, &ts.BriefDescription); err != nil {
		return nil, fmt.Errorf("services detail for %s not found: %w", txID, err)
	}
	return &ts, nil
}

// GetMilestonesByTxID returns the milestone breakdown of a services transaction.
func (r *TransactionRepository) GetMilestonesByTxID(ctx context.Context, txID string) ([]domain.ServiceMilestone, error) {
	query := `
		SELECT id, transaction_id, milestone_index, title, amount, status, released_at, created_at
		FROM service_milestones WHERE transaction_id = $1 ORDER BY milestone_index ASC
	`
	var rows *sql.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query, txID)
	} else {
		rows, err = r.db.QueryContext(ctx, query, txID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list milestones: %w", err)
	}
	defer rows.Close()

	out := []domain.ServiceMilestone{}
	for rows.Next() {
		var m domain.ServiceMilestone
		if err := rows.Scan(&m.ID, &m.TransactionID, &m.MilestoneIndex, &m.Title, &m.Amount, &m.Status, &m.ReleasedAt, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan milestone row: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetEventsDetailByTxID fetches the event detail container of a transaction.
func (r *TransactionRepository) GetEventsDetailByTxID(ctx context.Context, txID string) (*domain.TransactionEvents, error) {
	query := `
		SELECT transaction_id, sub_sub_category_id, event_name, event_start_time, event_end_time, ticket_quantity_total
		FROM transaction_events WHERE transaction_id = $1
	`
	var te domain.TransactionEvents
	row := func() interface{ Scan(...interface{}) error } {
		if r.tx != nil {
			return r.tx.QueryRowContext(ctx, query, txID)
		}
		return r.db.QueryRowContext(ctx, query, txID)
	}
	if err := row().Scan(
		&te.TransactionID, &te.SubSubCategoryID, &te.EventName, &te.EventStartTime, &te.EventEndTime, &te.TicketQuantityTotal,
	); err != nil {
		return nil, fmt.Errorf("event detail for %s not found: %w", txID, err)
	}
	return &te, nil
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
