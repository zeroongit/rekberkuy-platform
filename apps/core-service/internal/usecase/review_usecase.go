package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"rekberkuy/core-service/internal/domain"
)

// ReviewUsecase handles buyer→counterparty reviews. A review is only accepted
// on a Transaction that is RELEASED, and only from its buyer; the reviewee is
// the seller/service provider. Uniqueness (one per transaction) is enforced by
// the repository's composite index.
type ReviewUsecase struct {
	txRepo domain.TransactionRepository
	repo   domain.ReviewRepository
}

func NewReviewUsecase(txRepo domain.TransactionRepository, repo domain.ReviewRepository) *ReviewUsecase {
	return &ReviewUsecase{txRepo: txRepo, repo: repo}
}

func (u *ReviewUsecase) CreateReview(ctx context.Context, reviewerID, transactionID string, rating int, comment *string) (*domain.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	tx, err := u.txRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load transaction: %w", err)
	}
	if tx.Status != domain.StatusReleased {
		return nil, fmt.Errorf("cannot review a transaction that is not RELEASED (current: %s)", tx.Status)
	}
	if tx.BuyerID != reviewerID {
		return nil, errors.New("only the buyer of a transaction may leave a review")
	}

	review := &domain.Review{
		ID:            uuid.New().String(),
		TransactionID: transactionID,
		ReviewerID:    reviewerID,
		RevieweeID:    tx.SellerID,
		Rating:        rating,
		Comment:       comment,
	}
	if err := u.repo.CreateReview(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to save review: %w", err)
	}
	return review, nil
}

func (u *ReviewUsecase) GetReviewsForUser(ctx context.Context, userID string) ([]domain.Review, error) {
	return u.repo.GetReviewsByReviewee(ctx, userID)
}
