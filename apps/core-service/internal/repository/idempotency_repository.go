package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type idempotencyRepository struct {
	db *sql.DB
}

func NewIdempotencyRepository(db *sql.DB) *idempotencyRepository {
	return &idempotencyRepository{db: db}
}

// CheckOrLock checks for the presence of a key atomically using
// INSERT ... ON CONFLICT DO NOTHING. Race conditions between concurrent requests
// are prevented because PostgreSQL handles the UNIQUE conflict at the row level.
//
// Returns:
//   - (recordExisting, false, nil) if the key already exists (duplicate transaction)
//   - (rec, true, nil)              if the new insert succeeds (secures the request)
//   - (nil, false, err)             if an unexpected error occurs
func (r *idempotencyRepository) CheckOrLock(ctx context.Context, rec *domain.IdempotencyRecord) (*domain.IdempotencyRecord, bool, error) {
	insertQuery := `
		INSERT INTO idempotency_records (id, request_path, response_body, response_status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (id) DO NOTHING
	`
	res, err := r.db.ExecContext(ctx, insertQuery, rec.ID, rec.RequestPath, rec.ResponseBody, rec.ResponseStatus)
	if err != nil {
		return nil, false, fmt.Errorf("failed to lock idempotency key: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("failed to read idempotency insert result: %w", err)
	}

	// rowsAffected == 1 -> new insert succeeded (new request)
	if rowsAffected == 1 {
		return rec, true, nil
	}

	// rowsAffected == 0 -> conflict (key already exists) -> fetch the existing record
	existing, err := r.getExisting(ctx, rec.ID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch existing idempotency record: %w", err)
	}
	return existing, false, nil
}

// getExisting fetches a previously stored idempotency record.
func (r *idempotencyRepository) getExisting(ctx context.Context, id string) (*domain.IdempotencyRecord, error) {
	query := `SELECT id, request_path, response_body, response_status, created_at FROM idempotency_records WHERE id = $1`
	var existing domain.IdempotencyRecord
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&existing.ID, &existing.RequestPath, &existing.ResponseBody, &existing.ResponseStatus, &existing.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("inconsistency: key %s rejected on insert but not found on read", id)
		}
		return nil, err
	}
	return &existing, nil
}

// SaveResponse updates the idempotency record with the actual response from the first request.
func (r *idempotencyRepository) SaveResponse(ctx context.Context, id string, status int, body []byte) error {
	query := `UPDATE idempotency_records SET response_body = $1, response_status = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, body, status, id)
	if err != nil {
		return fmt.Errorf("failed to save idempotency response: %w", err)
	}
	return nil
}
