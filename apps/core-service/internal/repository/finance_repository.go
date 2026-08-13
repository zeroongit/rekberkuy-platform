package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type financeRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewFinanceRepository initializes the database adapter for platform finance
func NewFinanceRepository(db *sql.DB) *financeRepository {
	return &financeRepository{db: db}
}

// GetPlatformFinance fetches the global cash summary of the RekberKuy platform
func (r *financeRepository) GetPlatformFinance(ctx context.Context) (*domain.PlatformFinance, error) {
	query := `
		SELECT id, total_escrow_balance, total_revenue, total_midtrans_fees, updated_at
		FROM platform_finances
		LIMIT 1
	`
	var f domain.PlatformFinance
	var err error
	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query).Scan(&f.ID, &f.TotalEscrowBalance, &f.TotalRevenue, &f.TotalMidtransFees, &f.UpdatedAt)
	} else {
		err = r.db.QueryRowContext(ctx, query).Scan(&f.ID, &f.TotalEscrowBalance, &f.TotalRevenue, &f.TotalMidtransFees, &f.UpdatedAt)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			// If empty, return a default object with zero balance
			return &domain.PlatformFinance{}, nil
		}
		return nil, fmt.Errorf("failed to fetch platform global cash data: %w", err)
	}
	return &f, nil
}

// UpdatePlatformFinance updates the global cash atomically using deltas (additions/subtractions)
func (r *financeRepository) UpdatePlatformFinance(ctx context.Context, escrowDelta int64, revenueDelta int64, midtransFeeDelta int64) error {
	query := `
		INSERT INTO platform_finances (id, total_escrow_balance, total_revenue, total_midtrans_fees, updated_at)
		VALUES ('GLOBAL_FINANCE_ID', $1, $2, $3, NOW())
		ON CONFLICT (id) DO UPDATE SET
			total_escrow_balance = platform_finances.total_escrow_balance + EXCLUDED.total_escrow_balance,
			total_revenue = platform_finances.total_revenue + EXCLUDED.total_revenue,
			total_midtrans_fees = platform_finances.total_midtrans_fees + EXCLUDED.total_midtrans_fees,
			updated_at = NOW()
	`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, escrowDelta, revenueDelta, midtransFeeDelta)
	} else {
		_, err = r.db.ExecContext(ctx, query, escrowDelta, revenueDelta, midtransFeeDelta)
	}
	if err != nil {
		return fmt.Errorf("failed to update platform global cash mutations: %w", err)
	}
	return nil
}
