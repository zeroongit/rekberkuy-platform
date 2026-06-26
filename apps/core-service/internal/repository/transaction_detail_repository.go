package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type transactionDetailRepository struct {
	db *sql.DB
}

// NewTransactionDetailRepository menginisialisasi adapter untuk detail Jasa dan Event
func NewTransactionDetailRepository(db *sql.DB) *transactionDetailRepository {
	return &transactionDetailRepository{db: db}
}

// GetServicesDetailByID mengambil data deadline proyek dan deskripsi brief jasa
func (r *transactionDetailRepository) GetServicesDetailByID(ctx context.Context, txID string) (*domain.TransactionServices, error) {
	query := `
		SELECT transaction_id, project_deadline, brief_description 
		FROM transaction_services 
		WHERE transaction_id = $1
	`
	var serviceTx domain.TransactionServices
	err := r.db.QueryRowContext(ctx, query, txID).Scan(
		&serviceTx.TransactionID,
		&serviceTx.ProjectDeadline,
		&serviceTx.BriefDescription,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail transaksi jasa: %w", err)
	}
	return &serviceTx, nil
}

// CreateServiceMilestone mencatat termin/cicilan pembayaran baru untuk transaksi Jasa
func (r *transactionDetailRepository) CreateServiceMilestone(ctx context.Context, milestone *domain.ServiceMilestone) error {
	query := `
		INSERT INTO service_milestones (
			id, transaction_id, milestone_index, title, amount, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		milestone.ID,
		milestone.TransactionID,
		milestone.MilestoneIndex,
		milestone.Title,
		milestone.Amount,
		milestone.Status,
	)
	if err != nil {
		return fmt.Errorf("gagal mencatat milestone jasa: %w", err)
	}
	return nil
}