package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

type kycRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewKYCRepository(db *sql.DB) domain.KYCRepository {
	return &kycRepository{db: db}
}

const kycSubmissionColumns = `id, user_id, target_role, id_card_number, id_card_url, selfie_url,
	status, ai_score, ai_reason, admin_notes, reviewed_by, reviewed_at, created_at, updated_at`

func (r *kycRepository) exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if r.tx != nil {
		return r.tx.ExecContext(ctx, query, args...)
	}
	return r.db.ExecContext(ctx, query, args...)
}

func (r *kycRepository) queryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if r.tx != nil {
		return r.tx.QueryRowContext(ctx, query, args...)
	}
	return r.db.QueryRowContext(ctx, query, args...)
}

func (r *kycRepository) query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if r.tx != nil {
		return r.tx.QueryContext(ctx, query, args...)
	}
	return r.db.QueryContext(ctx, query, args...)
}

func (r *kycRepository) SubmitKYC(ctx context.Context, kyc *domain.KYCSubmission) error {
	query := `
		INSERT INTO kyc_submissions (id, user_id, target_role, id_card_number, id_card_url, selfie_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			target_role = EXCLUDED.target_role,
			id_card_number = EXCLUDED.id_card_number,
			id_card_url = EXCLUDED.id_card_url,
			selfie_url = EXCLUDED.selfie_url,
			status = 'PENDING',
			ai_score = NULL,
			ai_reason = NULL,
			admin_notes = NULL,
			reviewed_by = NULL,
			reviewed_at = NULL,
			updated_at = NOW()
	`
	_, err := r.exec(ctx, query,
		kyc.ID,
		kyc.UserID,
		string(kyc.TargetRole),
		kyc.IDCardNumber,
		kyc.IDCardURL,
		kyc.SelfieURL,
		string(kyc.Status),
	)
	if err != nil {
		return fmt.Errorf("failed to record kyc documents to database: %w", err)
	}
	return nil
}

func scanKYCSubmission(row *sql.Row, k *domain.KYCSubmission) error {
	return row.Scan(
		&k.ID, &k.UserID, &k.TargetRole, &k.IDCardNumber, &k.IDCardURL, &k.SelfieURL,
		&k.Status, &k.AIScore, &k.AIReason, &k.AdminNotes, &k.ReviewedBy, &k.ReviewedAt,
		&k.CreatedAt, &k.UpdatedAt,
	)
}

func (r *kycRepository) GetKYCByID(ctx context.Context, id string) (*domain.KYCSubmission, error) {
	query := `SELECT ` + kycSubmissionColumns + ` FROM kyc_submissions WHERE id = $1 FOR UPDATE`
	var k domain.KYCSubmission
	if err := scanKYCSubmission(r.queryRow(ctx, query, id), &k); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("kyc submission %s not found", id)
		}
		return nil, fmt.Errorf("failed to find kyc submission: %w", err)
	}
	return &k, nil
}

func (r *kycRepository) ListPendingKYCs(ctx context.Context) ([]domain.KYCSubmission, error) {
	query := `SELECT ` + kycSubmissionColumns + ` FROM kyc_submissions WHERE status = 'PENDING' ORDER BY created_at ASC`
	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending kyc submissions: %w", err)
	}
	defer rows.Close()

	out := []domain.KYCSubmission{}
	for rows.Next() {
		var k domain.KYCSubmission
		if err := rows.Scan(
			&k.ID, &k.UserID, &k.TargetRole, &k.IDCardNumber, &k.IDCardURL, &k.SelfieURL,
			&k.Status, &k.AIScore, &k.AIReason, &k.AdminNotes, &k.ReviewedBy, &k.ReviewedAt,
			&k.CreatedAt, &k.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan kyc submission row: %w", err)
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *kycRepository) SaveKYCAIResult(ctx context.Context, userID string, score float64, reason string) error {
	query := `UPDATE kyc_submissions SET ai_score = $2, ai_reason = $3, updated_at = NOW() WHERE user_id = $1`
	if _, err := r.exec(ctx, query, userID, score, reason); err != nil {
		return fmt.Errorf("failed to save kyc ai reference: %w", err)
	}
	return nil
}

func (r *kycRepository) ReviewKYC(ctx context.Context, id string, status domain.KYCStatus, adminID string, notes string) error {
	query := `
		UPDATE kyc_submissions
		SET status = $2, admin_notes = $3, reviewed_by = $4, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	if _, err := r.exec(ctx, query, id, string(status), notes, adminID); err != nil {
		return fmt.Errorf("failed to record kyc review decision: %w", err)
	}
	return nil
}
