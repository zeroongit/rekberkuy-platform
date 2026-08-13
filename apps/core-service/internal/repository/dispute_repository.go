package repository

import (
	"context"
	"database/sql"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

type disputeRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewDisputeRepository(db *sql.DB) domain.DisputeRepository {
	return &disputeRepository{db: db}
}

func (r *disputeRepository) CreateDispute(ctx context.Context, d *domain.Dispute) error {
	query := `
		INSERT INTO disputes (id, transaction_id, raised_by, target_party_id, reason, evidence_url, status, is_resolved, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	args := []any{d.ID, d.TransactionID, d.RaisedBy, d.TargetPartyID, d.Reason, d.EvidenceURL, d.Status, d.IsResolved}
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, args...)
	} else {
		_, err = r.db.ExecContext(ctx, query, args...)
	}
	if err != nil {
		return fmt.Errorf("failed to create dispute: %w", err)
	}
	return nil
}

const disputeColumns = `id, transaction_id, raised_by, target_party_id, mediator_id, reason, evidence_url, status, outcome, is_resolved, resolution_summary, resolved_at, created_at, updated_at`

func scanDispute(row interface {
	Scan(dest ...any) error
}, d *domain.Dispute) error {
	return row.Scan(
		&d.ID, &d.TransactionID, &d.RaisedBy, &d.TargetPartyID, &d.MediatorID, &d.Reason, &d.EvidenceURL,
		&d.Status, &d.Outcome, &d.IsResolved, &d.ResolutionSummary, &d.ResolvedAt, &d.CreatedAt, &d.UpdatedAt,
	)
}

func (r *disputeRepository) GetDisputeByID(ctx context.Context, id string) (*domain.Dispute, error) {
	query := `SELECT ` + disputeColumns + ` FROM disputes WHERE id = $1`
	if r.tx != nil {
		query += " FOR UPDATE"
	}
	var d domain.Dispute
	var err error
	if r.tx != nil {
		err = scanDispute(r.tx.QueryRowContext(ctx, query, id), &d)
	} else {
		err = scanDispute(r.db.QueryRowContext(ctx, query, id), &d)
	}
	if err != nil {
		return nil, fmt.Errorf("dispute %s not found: %w", id, err)
	}
	return &d, nil
}

func (r *disputeRepository) GetDisputeByTransactionID(ctx context.Context, transactionID string) (*domain.Dispute, error) {
	query := `SELECT ` + disputeColumns + ` FROM disputes WHERE transaction_id = $1`
	if r.tx != nil {
		query += " FOR UPDATE"
	}
	var d domain.Dispute
	var err error
	if r.tx != nil {
		err = scanDispute(r.tx.QueryRowContext(ctx, query, transactionID), &d)
	} else {
		err = scanDispute(r.db.QueryRowContext(ctx, query, transactionID), &d)
	}
	if err != nil {
		return nil, fmt.Errorf("dispute for transaction %s not found: %w", transactionID, err)
	}
	return &d, nil
}

func (r *disputeRepository) AcknowledgeDispute(ctx context.Context, id string) error {
	query := `UPDATE disputes SET status = $1, updated_at = NOW() WHERE id = $2`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, domain.DisputeStatusUnderReview, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, domain.DisputeStatusUnderReview, id)
	}
	if err != nil {
		return fmt.Errorf("failed to acknowledge dispute: %w", err)
	}
	return nil
}

func (r *disputeRepository) ResolveDispute(ctx context.Context, id string, adminID string, outcome domain.DisputeOutcome, summary string) error {
	query := `UPDATE disputes SET status = $1, outcome = $2, mediator_id = $3, resolution_summary = $4, is_resolved = TRUE, resolved_at = NOW(), updated_at = NOW() WHERE id = $5`
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, domain.DisputeStatusResolved, outcome, adminID, summary, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, domain.DisputeStatusResolved, outcome, adminID, summary, id)
	}
	if err != nil {
		return fmt.Errorf("failed to resolve dispute: %w", err)
	}
	return nil
}
