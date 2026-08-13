package usecase_test

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// DISBURSEMENT USECASE — UNIT TESTS
// ============================================================================

func TestDisbursement_MarkExternalPayoutDisbursed_Success(t *testing.T) {
	ctx := context.Background()
	marked := false
	txRepo := &mockTransactionRepo{
		onGetEventVendorPayoutByID: func(ctx context.Context, id string) (*domain.EventVendorPayout, error) {
			return &domain.EventVendorPayout{ID: "pay-1", Status: domain.VendorPayoutPendingDisbursement, VendorName: "Vendor Luar"}, nil
		},
		onMarkEventVendorPayoutDisbursed: func(ctx context.Context, payoutID, adminID string) error {
			marked = true
			if payoutID != "pay-1" || adminID != "admin-1" {
				t.Errorf("got payout=%s admin=%s, want pay-1/admin-1", payoutID, adminID)
			}
			return nil
		},
	}
	u := usecase.NewDisbursementUsecase(txRepo)

	if err := u.MarkExternalPayoutDisbursed(ctx, "pay-1", "admin-1"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !marked {
		t.Error("payout was not marked disbursed")
	}
}

func TestDisbursement_MarkExternalPayoutDisbursed_NotPendingDisbursement(t *testing.T) {
	txRepo := &mockTransactionRepo{
		onGetEventVendorPayoutByID: func(ctx context.Context, id string) (*domain.EventVendorPayout, error) {
			return &domain.EventVendorPayout{ID: "pay-1", Status: domain.VendorPayoutApproved}, nil // internal vendor / already approved
		},
		onMarkEventVendorPayoutDisbursed: func(ctx context.Context, payoutID, adminID string) error {
			t.Error("must not mark a non-PENDING_DISBURSEMENT payout")
			return nil
		},
	}
	u := usecase.NewDisbursementUsecase(txRepo)

	if err := u.MarkExternalPayoutDisbursed(context.Background(), "pay-1", "admin-1"); err == nil {
		t.Fatal("expected error because payout is not PENDING_DISBURSEMENT")
	}
}

func TestDisbursement_MarkExternalPayoutDisbursed_AlreadyDisbursed(t *testing.T) {
	txRepo := &mockTransactionRepo{
		onGetEventVendorPayoutByID: func(ctx context.Context, id string) (*domain.EventVendorPayout, error) {
			return &domain.EventVendorPayout{ID: "pay-1", Status: domain.VendorPayoutDisbursed}, nil
		},
		onMarkEventVendorPayoutDisbursed: func(ctx context.Context, payoutID, adminID string) error {
			t.Error("must not re-mark an already-disbursed payout")
			return nil
		},
	}
	u := usecase.NewDisbursementUsecase(txRepo)

	if err := u.MarkExternalPayoutDisbursed(context.Background(), "pay-1", "admin-1"); err == nil {
		t.Fatal("expected error because payout is already DISBURSED")
	}
}
