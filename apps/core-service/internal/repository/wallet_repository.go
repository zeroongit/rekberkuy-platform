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

// NewWalletRepository initializes the database adapter for the RekberPay Wallet
func NewWalletRepository(db *sql.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

// GetBalance fetches the balance and frozen status of a user's wallet from Supabase
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
			return nil, fmt.Errorf("wallet not found for user id: %s", userID)
		}
		return nil, err
	}

	return &wallet, nil
}

// CreateWallet creates a new RekberPay wallet when a user completes registration
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
		return fmt.Errorf("failed to create new wallet: %w", err)
	}

	return nil
}

// UpdateBalanceTx executes balance mutations safely with race-condition protection
func (r *walletRepository) UpdateBalanceTx(ctx context.Context, txRecord *domain.RekberPayTransaction, amountModifier int64) error {
	var tx *sql.Tx
	var err error
	isNestedTx := r.tx != nil

	if isNestedTx {
		tx = r.tx
	} else {
		tx, err = r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return fmt.Errorf("failed to begin database transaction: %w", err)
		}
	}

	defer func() {
		if err != nil && !isNestedTx {
			_ = tx.Rollback()
		}
	}()

	// 1. Lock the user's wallet row (SELECT FOR UPDATE)
	var currentBalance int64
	var isFrozen bool
	lockQuery := `
		SELECT balance, is_frozen FROM rekberpay_wallets WHERE user_id = $1 FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, lockQuery, txRecord.WalletID).Scan(&currentBalance, &isFrozen)
	if err != nil {
		return fmt.Errorf("failed to lock wallet data for mutation: %w", err)
	}

	if isFrozen {
		return fmt.Errorf("transaction rejected: wallet for user %s is currently frozen", txRecord.WalletID)
	}

	// 2. Calculate the balance deduction portion along with the flat Rp7.500 withdrawal fee
	totalDeduction := amountModifier
	if amountModifier < 0 {
		if txRecord.Type == domain.TxWithdraw {
			totalDeduction = amountModifier - txRecord.AdminFee
		}
		if (currentBalance + totalDeduction) < 0 {
			return errors.New("transaction rejected: RekberPay balance is not sufficient for the transaction amount plus admin fee")
		}
	}

	// 3. Update the master balance table
	updateWalletQuery := `
		UPDATE rekberpay_wallets SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2
	`
	_, err = tx.ExecContext(ctx, updateWalletQuery, totalDeduction, txRecord.WalletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// 4. Record the financial mutation history into the rekberpay_transactions table
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
		return fmt.Errorf("failed to record transaction mutation history: %w", err)
	}

	if !isNestedTx {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit financial transaction: %w", err)
		}
	}

	return nil
}

// GetAllUsers ForCRMEvaluation fetches all user profile data
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
		return nil, fmt.Errorf("failed to fetch CRM evaluation profile data: %w", err)
	}
	defer rows.Close()

	var users []*domain.UserProfile
	for rows.Next() {
		var user domain.UserProfile
		err := rows.Scan(
			&user.ID, &user.Username, &user.FullName, &user.Role, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user data: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetCRMLoyaltyByUserID fetches the user's loyalty tier profile
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
			return nil, fmt.Errorf("crm profile not found for user id: %s", userID)
		}
		return nil, err
	}

	return &crm, nil
}

// UpdateCRMLoyalty stores the updated tier resulting from the monthly evaluation
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
		return fmt.Errorf("failed to update user crm tier data: %w", err)
	}

	return nil
}

// GetVendorAllocationsByTxID fetches the list of vendor fund-allocation commitments as a pointer slice []*domain.EventVendorAllocation
func (r *walletRepository) GetVendorAllocationsByTxID(ctx context.Context, transactionID string) ([]*domain.EventVendorAllocation, error) {
	query := `
		SELECT id, transaction_id, vendor_id, allocated_amount, actual_paid_amount, status, created_at
		FROM event_vendor_allocations
		WHERE transaction_id = $1
	`
	var rows *sql.Rows
	var err error

	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query, transactionID)
	} else {
		rows, err = r.db.QueryContext(ctx, query, transactionID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query vendor allocations: %w", err)
	}
	defer rows.Close()

	var allocations []*domain.EventVendorAllocation
	for rows.Next() {
		var alloc domain.EventVendorAllocation
		err := rows.Scan(
			&alloc.ID,
			&alloc.TransactionID,
			&alloc.VendorID,
			&alloc.AllocatedAmount,
			&alloc.ActualPaidAmount,
			&alloc.Status,
			&alloc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vendor allocation row: %w", err)
		}
		allocations = append(allocations, &alloc) // 👈 Take its pointer (&alloc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return allocations, nil
}

// CreateVendorPayoutRecord records the official sub-vendor payment claim invoice to the database
func (r *walletRepository) CreateVendorPayoutRecord(ctx context.Context, payout *domain.EventVendorPayout) error {
	query := `
		INSERT INTO event_vendor_payouts (
			id, transaction_id, vendor_user_id, vendor_name, vendor_bank_name, vendor_account_number,
			amount_requested, expense_description, invoice_file_url, payout_phase, status,
			is_disbursed_by_midtrans, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query,
			payout.ID, payout.TransactionID, payout.VendorUserID, payout.VendorName, payout.VendorBankName, payout.VendorAccountNumber,
			payout.AmountRequested, payout.ExpenseDescription, payout.InvoiceFileURL, payout.PayoutPhase, payout.Status,
			payout.IsDisbursedByMidtrans,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			payout.ID, payout.TransactionID, payout.VendorUserID, payout.VendorName, payout.VendorBankName, payout.VendorAccountNumber,
			payout.AmountRequested, payout.ExpenseDescription, payout.InvoiceFileURL, payout.PayoutPhase, payout.Status,
			payout.IsDisbursedByMidtrans,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to save vendor payout claim data: %w", err)
	}
	return nil
}

// GetVendorPayoutByID fetches the event installment invoice request detail by Payout ID
func (r *walletRepository) GetVendorPayoutByID(ctx context.Context, payoutID string) (*domain.EventVendorPayout, error) {
	query := `
		SELECT
			id, transaction_id, vendor_user_id, vendor_name, vendor_bank_name, vendor_account_number,
			amount_requested, expense_description, invoice_file_url, payout_phase, status,
			is_disbursed_by_midtrans, created_at
		FROM event_vendor_payouts
		WHERE id = $1
	`
	var payout domain.EventVendorPayout
	var err error

	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, payoutID).Scan(
			&payout.ID, &payout.TransactionID, &payout.VendorUserID, &payout.VendorName, &payout.VendorBankName, &payout.VendorAccountNumber,
			&payout.AmountRequested, &payout.ExpenseDescription, &payout.InvoiceFileURL, &payout.PayoutPhase, &payout.Status,
			&payout.IsDisbursedByMidtrans, &payout.CreatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, payoutID).Scan(
			&payout.ID, &payout.TransactionID, &payout.VendorUserID, &payout.VendorName, &payout.VendorBankName, &payout.VendorAccountNumber,
			&payout.AmountRequested, &payout.ExpenseDescription, &payout.InvoiceFileURL, &payout.PayoutPhase, &payout.Status,
			&payout.IsDisbursedByMidtrans, &payout.CreatedAt,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch vendor payout request data: %w", err)
	}
	return &payout, nil
}

// UpdateVendorPayoutStatus updates the invoice approval status within an ACID scope
func (r *walletRepository) UpdateVendorPayoutStatus(ctx context.Context, payoutID string, status string) error {
	query := `
		UPDATE event_vendor_payouts
		SET status = $1
		WHERE id = $2
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, status, payoutID)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, payoutID)
	}

	if err != nil {
		return fmt.Errorf("failed to update vendor payout request status: %w", err)
	}
	return nil
}

// GetWalletTxByMidtransOrderID fetches the top-up draft transaction by its Midtrans order id.
func (r *walletRepository) GetWalletTxByMidtransOrderID(ctx context.Context, orderID string) (*domain.RekberPayTransaction, error) {
	query := `
		SELECT id, wallet_id, type, status, amount, admin_fee, platform_net_profit,
		       reference_transaction_id, midtrans_topup_id, description, created_at
		FROM rekberpay_transactions
		WHERE midtrans_topup_id = $1
	`
	// Lock the row inside a transaction so concurrent webhook retries cannot both
	// read PENDING and double-credit the wallet (anti double-spending).
	if r.tx != nil {
		query += " FOR UPDATE"
	}
	var t domain.RekberPayTransaction
	var err error
	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, orderID).Scan(
			&t.ID, &t.WalletID, &t.Type, &t.Status, &t.Amount, &t.AdminFee, &t.PlatformNetProfit,
			&t.ReferenceTransactionID, &t.MidtransTopUpID, &t.Description, &t.CreatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, orderID).Scan(
			&t.ID, &t.WalletID, &t.Type, &t.Status, &t.Amount, &t.AdminFee, &t.PlatformNetProfit,
			&t.ReferenceTransactionID, &t.MidtransTopUpID, &t.Description, &t.CreatedAt,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("top-up draft with order id %s not found: %w", orderID, err)
	}
	return &t, nil
}

// MarkWalletTxStatusByOrderID updates the status of the top-up transaction row.
func (r *walletRepository) MarkWalletTxStatusByOrderID(ctx context.Context, orderID string, status domain.WalletTxStatus) error {
	query := `UPDATE rekberpay_transactions SET status = $1 WHERE midtrans_topup_id = $2`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, status, orderID)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, orderID)
	}
	if err != nil {
		return fmt.Errorf("failed to update top-up status: %w", err)
	}
	return nil
}
