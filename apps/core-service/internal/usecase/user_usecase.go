package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"rekberkuy/core-service/internal/domain"
)

// UserUsecase owns wallet-facing operations that hang off a user context
// (Midtrans top-up draft + webhook confirmation). Profile creation is shared
// via initProfileAndWallet so the registration flow (AuthUsecase) and any
// future importer reuse the same atomic profile+wallet guarantee.
type UserUsecase struct {
	uow        domain.UnitOfWork
	userRepo   domain.UserRepository
	walletRepo domain.WalletRepository
	midtrans   domain.MidtransClient
}

func NewUserUsecase(uow domain.UnitOfWork, ur domain.UserRepository, wr domain.WalletRepository, midtrans domain.MidtransClient) *UserUsecase {
	return &UserUsecase{
		uow:        uow,
		userRepo:   ur,
		walletRepo: wr,
		midtrans:   midtrans,
	}
}

// initProfileAndWallet atomically creates a user profile and its zero-balance
// RekberPay wallet inside a UnitOfWork transaction. A server-generated UUID is
// assigned when the caller did not supply one (registration never does).
func initProfileAndWallet(ctx context.Context, uow domain.UnitOfWork, user *domain.UserProfile) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	return uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		if err := stores.Users.CreateProfile(ctx, user); err != nil {
			return fmt.Errorf("failed to save profile to database: %w", err)
		}

		// CreateWallet inserts a zero-balance wallet row (idempotent ON CONFLICT).
		// UpdateBalanceTx cannot be used here: it opens with SELECT ... FOR UPDATE
		// on a row that does not exist yet, which would error and abort registration.
		if err := stores.Wallets.CreateWallet(ctx, user.ID); err != nil {
			return fmt.Errorf("failed to create user RekberPay wallet: %w", err)
		}
		return nil
	})
}
