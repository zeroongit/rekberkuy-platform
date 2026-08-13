package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

func newAuthUsecase(userRepo domain.UserRepository, uow domain.UnitOfWork) *usecase.AuthUsecase {
	token := usecase.NewTokenService("test-secret", time.Hour)
	return usecase.NewAuthUsecase(uow, userRepo, token)
}

func TestAuth_Register_Success(t *testing.T) {
	ctx := context.Background()
	createCalled := false
	walletInitCalled := false
	userRepo := &mockUserRepo{
		onGetProfileByEmail: func(ctx context.Context, email string) (*domain.UserProfile, error) {
			return nil, errors.New("not found")
		},
		onCreateProfile: func(ctx context.Context, u *domain.UserProfile) error {
			createCalled = true
			if u.ID == "" {
				t.Error("user.ID must be generated")
			}
			if u.Email != "jane@example.com" {
				t.Errorf("email = %s, want jane@example.com", u.Email)
			}
			if u.PasswordHash == "" || u.PasswordHash == "password123" {
				t.Error("password must be hashed, not plaintext or empty")
			}
			return nil
		},
	}
	walletRepo := &mockWalletRepo{
		onCreateWallet: func(ctx context.Context, userID string) error {
			walletInitCalled = true
			if userID == "" {
				t.Error("wallet must be created for the new user id")
			}
			return nil
		},
	}
	uow := newMockUnitOfWork(nil, walletRepo, nil, userRepo)
	a := newAuthUsecase(userRepo, uow)

	user, err := a.Register(ctx, "jane@example.com", "jane", "password123", "Jane Doe")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !createCalled || !walletInitCalled {
		t.Fatal("profile + wallet must be created")
	}
	if user.Role != domain.RoleUser {
		t.Errorf("role = %s, want USER", user.Role)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
		t.Errorf("stored hash does not match password: %v", err)
	}
}

func TestAuth_Register_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		onGetProfileByEmail: func(ctx context.Context, email string) (*domain.UserProfile, error) {
			return &domain.UserProfile{ID: "existing", Email: email}, nil
		},
	}
	uow := newMockUnitOfWork(nil, &mockWalletRepo{}, nil, userRepo)
	a := newAuthUsecase(userRepo, uow)

	if _, err := a.Register(ctx, "dup@example.com", "dup", "password123", "Dup"); !errors.Is(err, usecase.ErrEmailAlreadyRegistered) {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got: %v", err)
	}
}

func TestAuth_Register_WeakPassword(t *testing.T) {
	a := newAuthUsecase(&mockUserRepo{}, newMockUnitOfWork(nil, &mockWalletRepo{}, nil, &mockUserRepo{}))
	if _, err := a.Register(context.Background(), "x@example.com", "x", "short", "X"); err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestAuth_Register_MissingFields(t *testing.T) {
	a := newAuthUsecase(&mockUserRepo{}, newMockUnitOfWork(nil, &mockWalletRepo{}, nil, &mockUserRepo{}))
	if _, err := a.Register(context.Background(), "", "x", "password123", "X"); err == nil {
		t.Fatal("expected error for missing email")
	}
}

func TestAuth_Login_Success(t *testing.T) {
	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	stored := &domain.UserProfile{ID: "u-1", Username: "john", Email: "john@example.com", PasswordHash: string(hash), Role: domain.RoleUser}
	userRepo := &mockUserRepo{
		onGetProfileByEmail: func(ctx context.Context, email string) (*domain.UserProfile, error) {
			return stored, nil
		},
	}
	a := newAuthUsecase(userRepo, newMockUnitOfWork(nil, &mockWalletRepo{}, nil, userRepo))

	token, user, err := a.Login(ctx, "john@example.com", "secret123")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if user.ID != "u-1" {
		t.Errorf("user id = %s, want u-1", user.ID)
	}
	if token == "" || !strings.Contains(token, ".") {
		t.Errorf("token looks invalid: %q", token)
	}
}

func TestAuth_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	stored := &domain.UserProfile{ID: "u-1", Email: "john@example.com", PasswordHash: string(hash)}
	userRepo := &mockUserRepo{
		onGetProfileByEmail: func(ctx context.Context, email string) (*domain.UserProfile, error) {
			return stored, nil
		},
	}
	a := newAuthUsecase(userRepo, newMockUnitOfWork(nil, &mockWalletRepo{}, nil, userRepo))

	_, _, err := a.Login(context.Background(), "john@example.com", "wrong-password")
	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuth_Login_UnknownEmail(t *testing.T) {
	userRepo := &mockUserRepo{
		onGetProfileByEmail: func(ctx context.Context, email string) (*domain.UserProfile, error) {
			return nil, errors.New("not found")
		},
	}
	a := newAuthUsecase(userRepo, newMockUnitOfWork(nil, &mockWalletRepo{}, nil, userRepo))

	if _, _, err := a.Login(context.Background(), "ghost@example.com", "whatever"); !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for unknown email, got: %v", err)
	}
}
