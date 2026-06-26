package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository menginisialisasi adapter database untuk profil pengguna
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

// CreateProfile menyimpan data user baru hasil registrasi ke Supabase
func (r *userRepository) CreateProfile(ctx context.Context, user *domain.UserProfile) error {
	query := `
		INSERT INTO user_profiles (
			id, username, full_name, role, phone_number, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.FullName,
		user.Role,
		user.PhoneNumber,
	)
	if err != nil {
		return fmt.Errorf("gagal menyimpan profil user baru: %w", err)
	}
	return nil
}

// GetProfileByID mengambil data profil berdasarkan UUID user
func (r *userRepository) GetProfileByID(ctx context.Context, id string) (*domain.UserProfile, error) {
	query := `
		SELECT id, username, full_name, role, phone_number, created_at, updated_at
		FROM user_profiles
		WHERE id = $1
	`
	var user domain.UserProfile
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.FullName,
		&user.Role,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal mencari profil user: %w", err)
	}
	return &user, nil
}