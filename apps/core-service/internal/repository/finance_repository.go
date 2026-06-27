package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type financeRepository struct {
	db *sql.DB
}

// NewFinanceRepository menginisialisasi adapter database untuk keuangan platform
func NewFinanceRepository(db *sql.DB) *financeRepository {
	return &financeRepository{db: db}
}

// GetPlatformFinance mengambil rangkuman kas global platform RekberKuy
func (r *financeRepository) GetPlatformFinance(ctx context.Context) (*domain.PlatformFinance, error) {
	query := `
		SELECT id, total_escrow_balance, total_revenue, total_midtrans_fees, updated_at
		FROM platform_finances
		LIMIT 1
	`
	var f domain.PlatformFinance
	err := r.db.QueryRowContext(ctx, query).Scan(
		&f.ID,
		&f.TotalEscrowBalance,
		&f.TotalRevenue,
		&f.TotalMidtransFees,
		&f.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Jika kosong, return objek default dengan saldo 0
			return &domain.PlatformFinance{}, nil
		}
		return nil, fmt.Errorf("gagal mengambil data kas global platform: %w", err)
	}
	return &f, nil
}

// UpdatePlatformFinance memperbarui kas global secara atomik menggunakan delta (penambahan/pengurangan)
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
	_, err := r.db.ExecContext(ctx, query, escrowDelta, revenueDelta, midtransFeeDelta)
	if err != nil {
		return fmt.Errorf("gagal memperbarui mutasi kas global platform: %w", err)
	}
	return nil
}