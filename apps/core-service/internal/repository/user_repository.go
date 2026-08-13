package repository

import (
	"context"
	"database/sql"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

type userRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewUserRepository initializes the database adapter for user profiles.
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

// CreateProfile stores a new user's data (including auth credentials) into the database.
func (r *userRepository) CreateProfile(ctx context.Context, user *domain.UserProfile) error {
	query := `
		INSERT INTO user_profiles (
			id, username, email, password_hash, full_name, role, phone_number, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	args := []any{user.ID, user.Username, user.Email, user.PasswordHash, user.FullName, user.Role, user.PhoneNumber}
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, args...)
	} else {
		_, err = r.db.ExecContext(ctx, query, args...)
	}
	if err != nil {
		return fmt.Errorf("failed to save new user profile: %w", err)
	}
	return nil
}

const userProfileColumns = `id, username, email, password_hash, full_name, role, phone_number, created_at, updated_at`

func scanUserProfile(row interface {
	Scan(dest ...any) error
}, user *domain.UserProfile) error {
	return row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.Role, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt,
	)
}

// GetProfileByID fetches a profile by the user's UUID.
func (r *userRepository) GetProfileByID(ctx context.Context, id string) (*domain.UserProfile, error) {
	query := `SELECT ` + userProfileColumns + ` FROM user_profiles WHERE id = $1`
	var user domain.UserProfile
	var err error
	if r.tx != nil {
		err = scanUserProfile(r.tx.QueryRowContext(ctx, query, id), &user)
	} else {
		err = scanUserProfile(r.db.QueryRowContext(ctx, query, id), &user)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user profile: %w", err)
	}
	return &user, nil
}

// GetProfileByEmail fetches a profile by its unique email (used by the login flow).
func (r *userRepository) GetProfileByEmail(ctx context.Context, email string) (*domain.UserProfile, error) {
	query := `SELECT ` + userProfileColumns + ` FROM user_profiles WHERE email = $1`
	var user domain.UserProfile
	var err error
	if r.tx != nil {
		err = scanUserProfile(r.tx.QueryRowContext(ctx, query, email), &user)
	} else {
		err = scanUserProfile(r.db.QueryRowContext(ctx, query, email), &user)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user profile by email: %w", err)
	}
	return &user, nil
}
