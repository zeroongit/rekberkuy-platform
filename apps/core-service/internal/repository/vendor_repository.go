package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type vendorRepository struct {
	db *sql.DB
}

func NewVendorRepository(db *sql.DB) domain.VendorRepository {
	return &vendorRepository{db: db}
}

func (r *vendorRepository) CreateVendor(ctx context.Context, vendor *domain.VendorProfile) error {
	query := `
		INSERT INTO vendor_profiles (vendor_id, business_name, category, is_verified, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (vendor_id) DO UPDATE SET
			business_name = EXCLUDED.business_name,
			category = EXCLUDED.category,
			is_verified = EXCLUDED.is_verified
	`
	_, err := r.db.ExecContext(ctx, query, vendor.VendorID, vendor.BusinessName, vendor.Category, vendor.IsVerified)
	if err != nil {
		return fmt.Errorf("failed to record vendor profile data to database: %w", err)
	}
	return nil
}

// ListVendors returns marketplace vendors, optionally filtered by category
// (e.g. "CATERING"), newest first.
func (r *vendorRepository) ListVendors(ctx context.Context, category string, limit, offset int) ([]domain.VendorProfile, error) {
	query := `
		SELECT vendor_id, business_name, category, is_verified, created_at
		FROM vendor_profiles
	`
	args := []any{}
	if category != "" {
		query += ` WHERE category = $1`
		args = append(args, category)
	}
	query += ` ORDER BY created_at DESC`
	argIdx := func(n int) string { return fmt.Sprintf("$%d", n) }
	query += fmt.Sprintf(` LIMIT %s OFFSET %s`, argIdx(len(args)+1), argIdx(len(args)+2))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendors: %w", err)
	}
	defer rows.Close()

	out := []domain.VendorProfile{}
	for rows.Next() {
		var v domain.VendorProfile
		if err := rows.Scan(&v.VendorID, &v.BusinessName, &v.Category, &v.IsVerified, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan vendor row: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
