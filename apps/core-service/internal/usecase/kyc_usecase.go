package usecase

import (
	"context"
	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

type KYCUsecase struct {
	kycRepo domain.KYCRepository
}

func NewKYCUsecase(kr domain.KYCRepository) *KYCUsecase {
	return &KYCUsecase{kycRepo: kr}
}

func (u *KYCUsecase) SubmitUserKYC(ctx context.Context, userID string, targetRole domain.UserRole, idCard string, idCardURL string, selfieURL string) error {
	kycData := &domain.KYCSubmission{
		ID:           uuid.New().String(),
		UserID:       userID,
		TargetRole:   targetRole,
		IDCardNumber: idCard,
		IDCardURL:    idCardURL,
		SelfieURL:    selfieURL,
		Status:       domain.KYCPending, // Otomatis berstatus PENDING sesuai enum domain asli
	}
	return u.kycRepo.SubmitKYC(ctx, kycData)
}