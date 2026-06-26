package usecase

import (
	"context"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type UserUsecase struct {
	userRepo   domain.UserRepository
	walletRepo domain.WalletRepository
}

func NewUserUsecase(ur domain.UserRepository, wr domain.WalletRepository) *UserUsecase {
	return &UserUsecase{
		userRepo:   ur,
		walletRepo: wr,
	}
}

func (u *UserUsecase) RegisterNewUserProfile(ctx context.Context, user *domain.UserProfile) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		

		err := u.userRepo.CreateProfile(ctx, user)
		if err != nil {
			return fmt.Errorf("gagal menyimpan profil ke database: %w", err)
		}
		

		descMsg := "Inisialisasi pembuatan dompet RekberPay awal sukses."
		initTx := &domain.RekberPayTransaction{
			WalletID:    user.ID,
			Type:        "TOP_UP",
			Status:      domain.WalletStatusSuccess,
			Amount:      0,
			Description: &descMsg,
		}

		err = txRepo.UpdateBalanceTx(ctx, initTx, 0)
		if err != nil {
			return fmt.Errorf("gagal melahirkan dompet RekberPay user: %w", err)
		}

		return nil
	})
}