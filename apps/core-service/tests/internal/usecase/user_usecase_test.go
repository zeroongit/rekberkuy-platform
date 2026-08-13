package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// USER USECASE (TopUpWallet + ConfirmTopUp) — UNIT TESTS
// Registration is covered in auth_usecase_test.go (AuthUsecase.Register).
// ============================================================================

// ---------------------------------------------------------------------------
// TopUpWallet
// ---------------------------------------------------------------------------

func TestTopUpWallet_Success(t *testing.T) {
	ctx := context.Background()

	var captured *domain.RekberPayTransaction
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			captured = rec
			if modifier != 0 {
				t.Errorf("draft top-up modifier must be 0, got %d", modifier)
			}
			return nil
		},
	}
	midtrans := &mockMidtransClient{}
	u := usecase.NewUserUsecase(newMockUnitOfWork(nil, walletRepo, nil, nil), &mockUserRepo{}, walletRepo, midtrans)

	draft, snap, err := u.TopUpWallet(ctx, "user-1", 150000)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if draft.Status != domain.WalletStatusPending {
		t.Errorf("status draft = %s, want PENDING", draft.Status)
	}
	if draft.Type != domain.TxTopUp {
		t.Errorf("type = %s, want TOPUP", draft.Type)
	}
	if captured == nil || captured.Amount != 150000 {
		t.Errorf("top-up amount not recorded correctly: %+v", captured)
	}
	if snap == nil || snap.Token == "" {
		t.Error("snap token must not be empty")
	}
	if draft.MidtransTopUpID == nil || !strings.HasPrefix(*draft.MidtransTopUpID, domain.OrderPrefixTopUp) {
		t.Error("top-up order id must be generated with the correct prefix")
	}
}

func TestTopUpWallet_InvalidAmount(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewUserUsecase(newMockUnitOfWork(nil, &mockWalletRepo{}, nil, nil), &mockUserRepo{}, &mockWalletRepo{}, &mockMidtransClient{})

	if _, _, err := u.TopUpWallet(ctx, "user-1", 0); err == nil {
		t.Fatal("expected error for top-up 0")
	}
	if _, _, err := u.TopUpWallet(ctx, "user-1", -5000); err == nil {
		t.Fatal("expected error for negative top-up")
	}
}

func TestTopUpWallet_MidtransNotConfigured(t *testing.T) {
	ctx := context.Background()
	// midtrans nil -> usecase must reject with a clear message.
	u := usecase.NewUserUsecase(newMockUnitOfWork(nil, &mockWalletRepo{}, nil, nil), &mockUserRepo{}, &mockWalletRepo{}, nil)
	if _, _, err := u.TopUpWallet(ctx, "user-1", 50000); err == nil {
		t.Fatal("expected error because midtrans not configured")
	}
}

func TestTopUpWallet_SnapFailureMarksDraftFailed(t *testing.T) {
	ctx := context.Background()
	markCalled := false
	walletRepo := &mockWalletRepo{
		onMarkWalletTxStatus: func(ctx context.Context, orderID string, status domain.WalletTxStatus) error {
			if status != domain.WalletStatusFailed {
				t.Errorf("status = %s, want FAILED", status)
			}
			markCalled = true
			return nil
		},
	}
	midtrans := &mockMidtransClient{
		onCreateSnap: func(ctx context.Context, req domain.SnapRequest) (domain.SnapResult, error) {
			return domain.SnapResult{}, errors.New("midtrans unavailable")
		},
	}
	u := usecase.NewUserUsecase(newMockUnitOfWork(nil, walletRepo, nil, nil), &mockUserRepo{}, walletRepo, midtrans)

	if _, _, err := u.TopUpWallet(ctx, "user-1", 50000); err == nil {
		t.Fatal("expected error when snap fails")
	}
	if !markCalled {
		t.Error("draft top-up must be marked FAILED when snap fails")
	}
}

// ---------------------------------------------------------------------------
// ConfirmTopUp
// ---------------------------------------------------------------------------

func TestConfirmTopUp_CreditsWallet(t *testing.T) {
	ctx := context.Background()
	orderID := "REKBERKUY-TOPUP-abcd1234"
	draft := &domain.RekberPayTransaction{
		WalletID:        "user-1",
		Status:          domain.WalletStatusPending,
		Amount:          200000,
		MidtransTopUpID: &orderID,
	}
	walletRepo := &mockWalletRepo{
		onGetWalletTxByOrderID: func(ctx context.Context, oid string) (*domain.RekberPayTransaction, error) {
			return draft, nil
		},
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			if modifier != 200000 {
				t.Errorf("top-up credit = %d, want 200000", modifier)
			}
			return nil
		},
	}
	u := usecase.NewUserUsecase(newMockUnitOfWork(nil, walletRepo, nil, nil), &mockUserRepo{}, walletRepo, nil)

	if err := u.ConfirmTopUp(ctx, orderID, "midtrans-tx-1"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestConfirmTopUp_Idempotent(t *testing.T) {
	ctx := context.Background()
	orderID := "REKBERKUY-TOPUP-idem1"
	alreadySuccess := &domain.RekberPayTransaction{
		WalletID:        "user-1",
		Status:          domain.WalletStatusSuccess,
		Amount:          200000,
		MidtransTopUpID: &orderID,
	}
	credited := false
	walletRepo := &mockWalletRepo{
		onGetWalletTxByOrderID: func(ctx context.Context, oid string) (*domain.RekberPayTransaction, error) {
			return alreadySuccess, nil
		},
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			credited = true
			return nil
		},
	}
	u := usecase.NewUserUsecase(newMockUnitOfWork(nil, walletRepo, nil, nil), &mockUserRepo{}, walletRepo, nil)

	if err := u.ConfirmTopUp(ctx, orderID, "midtrans-tx-2"); err != nil {
		t.Fatalf("expected success (idempotent), got: %v", err)
	}
	if credited {
		t.Error("wallet must not be double-credited when draft is already SUCCESS")
	}
}
