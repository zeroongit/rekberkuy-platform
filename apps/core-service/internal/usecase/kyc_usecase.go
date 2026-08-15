package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"rekberkuy/core-service/internal/domain"
)

var (
	ErrKYCAlreadyReviewed = errors.New("kyc submission has already been reviewed")
	ErrKYCInvalidDecision = errors.New("kyc review decision must be APPROVED or REJECTED")
)

type KYCUsecase struct {
	kycRepo   domain.KYCRepository
	uow       domain.UnitOfWork
	kycClient domain.KYCClient
}

func NewKYCUsecase(kr domain.KYCRepository, uow domain.UnitOfWork, kc domain.KYCClient) *KYCUsecase {
	return &KYCUsecase{kycRepo: kr, uow: uow, kycClient: kc}
}

// SubmitUserKYC records the submission as PENDING, then asks backend-ai for a
// verification reference (confidence score) that the reviewing admin will see.
// The AI is a VERIFICATION service only: an unreachable/erroring AI service
// never blocks the submission — the row simply stays without an AI reference
// and the admin still reviews the raw documents.
func (u *KYCUsecase) SubmitUserKYC(ctx context.Context, userID string, targetRole domain.UserRole, idCard string, idCardURL string, selfieURL string) error {
	kycData := &domain.KYCSubmission{
		ID:           uuid.New().String(),
		UserID:       userID,
		TargetRole:   targetRole,
		IDCardNumber: idCard,
		IDCardURL:    idCardURL,
		SelfieURL:    selfieURL,
		Status:       domain.KYCPending,
	}
	if err := u.kycRepo.SubmitKYC(ctx, kycData); err != nil {
		return err
	}

	score, reason, err := u.kycClient.VerifyIdentity(ctx, userID, idCardURL, selfieURL, targetRole)
	if err != nil {
		// Reference-only failure: the submission remains valid & PENDING.
		log.Printf("⚠️  KYC AI reference unavailable for user %s: %v", userID, err)
		return nil
	}
	return u.kycRepo.SaveKYCAIResult(ctx, userID, score, reason)
}

// GetKYCByID returns a single submission (admin review detail view).
func (u *KYCUsecase) GetKYCByID(ctx context.Context, id string) (*domain.KYCSubmission, error) {
	return u.kycRepo.GetKYCByID(ctx, id)
}

// ListPendingKYCs returns the admin's verification queue.
func (u *KYCUsecase) ListPendingKYCs(ctx context.Context) ([]domain.KYCSubmission, error) {
	return u.kycRepo.ListPendingKYCs(ctx)
}

// ReviewKYC records the ADMIN decision on a PENDING submission. On APPROVE the
// submitter's role is promoted to the requested target role — both mutations
// run in one UnitOfWork so a role can never be promoted while the submission
// stays PENDING (and vice versa).
func (u *KYCUsecase) ReviewKYC(ctx context.Context, adminID string, kycID string, decision domain.KYCStatus, notes string) (*domain.KYCSubmission, error) {
	if decision != domain.KYCApproved && decision != domain.KYCRejected {
		return nil, ErrKYCInvalidDecision
	}

	var reviewed *domain.KYCSubmission
	err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		k, err := stores.KYC.GetKYCByID(ctx, kycID)
		if err != nil {
			return err
		}
		if k.Status != domain.KYCPending {
			return ErrKYCAlreadyReviewed
		}

		if err := stores.KYC.ReviewKYC(ctx, kycID, decision, adminID, notes); err != nil {
			return err
		}

		if decision == domain.KYCApproved {
			if err := stores.Users.UpdateUserRole(ctx, k.UserID, k.TargetRole); err != nil {
				return fmt.Errorf("failed to promote user role after kyc approval: %w", err)
			}
		}

		k.Status = decision
		k.AdminNotes = &notes
		k.ReviewedBy = &adminID
		reviewed = k
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reviewed, nil
}
