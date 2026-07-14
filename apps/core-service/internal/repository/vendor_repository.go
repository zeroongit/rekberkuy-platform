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
		return fmt.Errorf("gagal mencatat data profil vendor ke database: %w", err)
	}
	return nil
}