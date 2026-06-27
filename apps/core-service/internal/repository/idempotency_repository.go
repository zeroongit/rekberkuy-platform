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

// CheckOrLock memeriksa apakah key sudah ada. Jika belum, data dikunci (save). 
// Mengembalikan data lama jika duplicate, dan boolean true jika aman (new insert).
func (r *idempotencyRepository) CheckOrLock(ctx context.Context, rec *domain.IdempotencyRecord) (*domain.IdempotencyRecord, bool, error) {
	// Kueri baca aman
	checkQuery := `SELECT id, request_path, response_body, response_status, created_at FROM idempotency_records WHERE id = $1`
	var existing domain.IdempotencyRecord
	err := r.db.QueryRowContext(ctx, checkQuery, rec.ID).Scan(
		&existing.ID, &existing.RequestPath, &existing.ResponseBody, &existing.ResponseStatus, &existing.CreatedAt,
	)
	
	if err == nil {
		// Data sudah ada sebelumnya (Transaksi Ganda Terdeteksi!)
		return &existing, false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("gagal cek idempotensi: %w", err)
	}

	// Jika belum ada, masukkan record baru ke database
	insertQuery := `
		INSERT INTO idempotency_records (id, request_path, response_body, response_status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err = r.db.ExecContext(ctx, insertQuery, rec.ID, rec.RequestPath, rec.ResponseBody, rec.ResponseStatus)
	if err != nil {
		return nil, false, fmt.Errorf("gagal mengunci kunci idempotensi: %w", err)
	}

	return rec, true, nil
}