package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"rekberkuy/core-service/internal/domain"
)

// Auth-specific sentinel errors. Handlers map these to HTTP statuses.
var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid email or password")
)

// Password policy.
const minPasswordLength = 8

// TokenService mints signed HS256 JWTs for authenticated sessions. It is the
// single source of truth for token signing so the dev token-test endpoint and
// the real login flow produce identical tokens.
type TokenService struct {
	secret   string
	lifetime time.Duration
}

// NewTokenService receives the JWT secret (already validated by config) and a
// token lifetime parsed from JWT_TOKEN_LIFETIME.
func NewTokenService(secret string, lifetime time.Duration) *TokenService {
	return &TokenService{secret: secret, lifetime: lifetime}
}

// GenerateToken issues a token carrying the user's identity and role.
func (t *TokenService) GenerateToken(user *domain.UserProfile) (string, error) {
	claims := &domain.JWTCustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.lifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.secret))
}

// AuthUsecase owns credential-based registration and login. Registration reuses
// the shared profile+wallet creation helper so the atomic guarantee is identical
// to the legacy flow.
type AuthUsecase struct {
	uow      domain.UnitOfWork
	userRepo domain.UserRepository
	token    *TokenService
}

func NewAuthUsecase(uow domain.UnitOfWork, ur domain.UserRepository, token *TokenService) *AuthUsecase {
	return &AuthUsecase{uow: uow, userRepo: ur, token: token}
}

// Register validates the input, enforces email uniqueness, hashes the password
// (bcrypt), and atomically creates the profile + zero-balance RekberPay wallet.
// Returns the created profile (PasswordHash is never JSON-serialized).
func (a *AuthUsecase) Register(ctx context.Context, email, username, password, fullName string) (*domain.UserProfile, error) {
	if email == "" || username == "" || password == "" || fullName == "" {
		return nil, errors.New("email, username, password, and full name are required")
	}
	if len(password) < minPasswordLength {
		return nil, fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	// Email uniqueness pre-check. A DB unique constraint still backs this up.
	if existing, err := a.userRepo.GetProfileByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.UserProfile{
		ID:           uuid.New().String(),
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         domain.RoleUser,
	}

	if err := initProfileAndWallet(ctx, a.uow, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login verifies credentials and returns a signed JWT plus the user profile.
// On any failure (unknown email or wrong password) it returns the same error
// to avoid leaking which one was wrong.
func (a *AuthUsecase) Login(ctx context.Context, email, password string) (string, *domain.UserProfile, error) {
	user, err := a.userRepo.GetProfileByEmail(ctx, email)
	if err != nil || user == nil {
		return "", nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}
	token, err := a.token.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}
	return token, user, nil
}
