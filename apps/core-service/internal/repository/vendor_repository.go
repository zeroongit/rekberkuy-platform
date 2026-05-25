package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

type vendorRepository struct {
	db *sql.DB
}

// NewVendorRepository menginisialisasi adapter database untuk manajemen Vendor
func NewVendorRepository(db *sql.DB) *vendorRepository {
	return &vendorRepository{db: db}
}

// GetVendorProfile mengambil profil lengkap vendor beserta status verifikasinya
func (r *vendorRepository) GetVendorProfile(ctx context.Context, vendorID string) (*domain.UserProfile, error) {
	query := `
		SELECT id, username, full_name, role, wallet_address, phone_number, created_at, updated_at
		FROM user_profiles
		WHERE id = $1 AND role = 'VERIFIED_VENDOR'
	`

	var profile domain.UserProfile
	err := r.db.QueryRowContext(ctx, query, vendorID).Scan(
		&profile.ID,
		&profile.Username,
		&profile.FullName,
		&profile.Role,
		&profile.PhoneNumber,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("vendor tidak ditemukan atau akun belum terverifikasi sebagai VERIFIED_VENDOR")
		}
		return nil, err
	}

	return &profile, nil
}

// VerifyVendorDetail mengecek apakah data rekening atau identitas vendor valid sebelum sistem mencairkan dana termin event
func (r *vendorRepository) IsVendorActive(ctx context.Context, vendorID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM user_profiles 
			WHERE id = $1 AND role = 'VERIFIED_VENDOR'
		)
	`
	var isActive bool
	err := r.db.QueryRowContext(ctx, query, vendorID).Scan(&isActive)
	if err != nil {
		return false, err
	}
	return isActive, nil
}

func (r *vendorRepository) GetOrCreateCategory(ctx context.Context, categoryName string) (int, error) {
	// Query murni PostgreSQL: Masukkan jika belum ada, jika konflik (sudah ada) cukup ambil data yang lama
	query := `
		INSERT INTO vendor_categories (name) 
		VALUES (UPPER(TRIM($1)))
		ON CONFLICT (name) 
		DO UPDATE SET name = EXCLUDED.name
		RETURNING id;
	`
	
	var categoryID int
	err := r.db.QueryRowContext(ctx, query, categoryName).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("gagal memproses kategori vendor secara dinamis: %w", err)
	}
	
	return categoryID, nil
}