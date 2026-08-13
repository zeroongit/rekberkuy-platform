package usecase_test

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// REVIEW USECASE — UNIT TESTS
// ============================================================================

func TestReview_CreateReview_Success(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusReleased}, nil
		},
	}
	created := false
	reviewRepo := &mockReviewRepo{
		onCreateReview: func(ctx context.Context, rv *domain.Review) error {
			created = true
			if rv.ReviewerID != "buyer-1" || rv.RevieweeID != "seller-1" {
				t.Errorf("reviewer=%s reviewee=%s, want buyer-1/seller-1", rv.ReviewerID, rv.RevieweeID)
			}
			if rv.ID == "" {
				t.Error("review ID must be generated")
			}
			return nil
		},
	}
	u := usecase.NewReviewUsecase(txRepo, reviewRepo)

	rv, err := u.CreateReview(ctx, "buyer-1", "tx-1", 5, nil)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !created {
		t.Error("CreateReview was not called")
	}
	if rv.RevieweeID != "seller-1" {
		t.Errorf("reviewee = %s, want seller-1", rv.RevieweeID)
	}
}

func TestReview_CreateReview_NotReleased(t *testing.T) {
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusFundsLocked}, nil
		},
	}
	u := usecase.NewReviewUsecase(txRepo, &mockReviewRepo{
		onCreateReview: func(ctx context.Context, rv *domain.Review) error {
			t.Error("must not create review on a non-released transaction")
			return nil
		},
	})
	if _, err := u.CreateReview(context.Background(), "buyer-1", "tx-1", 5, nil); err == nil {
		t.Fatal("expected error because transaction is not RELEASED")
	}
}

func TestReview_CreateReview_ReviewerNotBuyer(t *testing.T) {
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusReleased}, nil
		},
	}
	u := usecase.NewReviewUsecase(txRepo, &mockReviewRepo{})
	if _, err := u.CreateReview(context.Background(), "intruder", "tx-1", 5, nil); err == nil {
		t.Fatal("expected error because reviewer is not the buyer")
	}
}

func TestReview_CreateReview_InvalidRating(t *testing.T) {
	u := usecase.NewReviewUsecase(&mockTransactionRepo{}, &mockReviewRepo{})
	if _, err := u.CreateReview(context.Background(), "b", "tx-1", 0, nil); err == nil {
		t.Fatal("expected error for rating 0")
	}
	if _, err := u.CreateReview(context.Background(), "b", "tx-1", 6, nil); err == nil {
		t.Fatal("expected error for rating 6")
	}
}

func TestReview_GetReviewsForUser(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		onGetReviewsByReviewee: func(ctx context.Context, id string) ([]domain.Review, error) {
			if id != "seller-1" {
				t.Errorf("queried id=%s, want seller-1", id)
			}
			return []domain.Review{{ID: "r1", RevieweeID: id, Rating: 4}}, nil
		},
	}
	u := usecase.NewReviewUsecase(&mockTransactionRepo{}, reviewRepo)
	out, err := u.GetReviewsForUser(context.Background(), "seller-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(out) != 1 {
		t.Errorf("len=%d, want 1", len(out))
	}
}
