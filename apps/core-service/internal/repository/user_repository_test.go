package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"rekberkuy/core-service/internal/domain"
)

func TestUserRepository_CreateProfile(t *testing.T) {
	db, mock := newMockDB(t)
	r := &userRepository{db: db}
	mock.ExpectExec(`INSERT INTO user_profiles`).
		WithArgs("u-1", "john", "john@example.com", "hash", "John Doe", domain.RoleUser, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	user := &domain.UserProfile{ID: "u-1", Username: "john", Email: "john@example.com", PasswordHash: "hash", FullName: "John Doe", Role: domain.RoleUser}
	if err := r.CreateProfile(context.Background(), user); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestUserRepository_GetProfileByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &userRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "full_name", "role", "phone_number", "created_at", "updated_at"}).
		AddRow("u-1", "john", "john@example.com", "hash", "John Doe", "USER", "0812", now, now)
	mock.ExpectQuery(`FROM user_profiles WHERE id = \$1`).WithArgs("u-1").WillReturnRows(rows)
	u, err := r.GetProfileByID(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if u.Username != "john" {
		t.Errorf("username = %s, want john", u.Username)
	}
	if u.Email != "john@example.com" {
		t.Errorf("email = %s, want john@example.com", u.Email)
	}
}

func TestUserRepository_GetProfileByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &userRepository{db: db}
	mock.ExpectQuery(`FROM user_profiles`).WillReturnError(sql.ErrNoRows)
	if _, err := r.GetProfileByID(context.Background(), "ghost"); err == nil {
		t.Fatal("expected error because user not found")
	}
}

func TestUserRepository_GetProfileByEmail_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &userRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "full_name", "role", "phone_number", "created_at", "updated_at"}).
		AddRow("u-1", "john", "john@example.com", "hash", "John Doe", "USER", nil, now, now)
	mock.ExpectQuery(`FROM user_profiles WHERE email = \$1`).WithArgs("john@example.com").WillReturnRows(rows)
	u, err := r.GetProfileByEmail(context.Background(), "john@example.com")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if u.ID != "u-1" {
		t.Errorf("id = %s, want u-1", u.ID)
	}
	if u.PasswordHash != "hash" {
		t.Errorf("password_hash = %s, want hash", u.PasswordHash)
	}
}

func TestUserRepository_GetProfileByEmail_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &userRepository{db: db}
	mock.ExpectQuery(`FROM user_profiles WHERE email =`).WillReturnError(sql.ErrNoRows)
	if _, err := r.GetProfileByEmail(context.Background(), "ghost@example.com"); err == nil {
		t.Fatal("expected error because email not found")
	}
}
