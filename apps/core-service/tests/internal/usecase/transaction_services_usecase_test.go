package usecase_test

import (
	"context"
	"errors"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// TRANSACTION SERVICES USECASE — UNIT TESTS
// ============================================================================

func TestLockFundsServices_Success(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{}
	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	// Services: buyer fee = 0, seller fee = flat 5% -> amountGross=100000, amountNet=95000
	tx, err := u.LockFundsServices(ctx, "buyer-1", "seller-1", 100000, true, "BRONZE", "REKBERPAY", "idem-svc-1", validServicesDetail(), validServicesMilestones(100000))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if tx.Type != domain.TypeServices {
		t.Errorf("type = %s, want %s", tx.Type, domain.TypeServices)
	}
	if tx.Status != domain.StatusWaitingPayment {
		t.Errorf("status = %s, want %s", tx.Status, domain.StatusWaitingPayment)
	}
	if tx.AmountGross != 100000 {
		t.Errorf("AmountGross = %d, want 100000 (services buyer fee 0)", tx.AmountGross)
	}
	if tx.AmountNet != 95000 {
		t.Errorf("AmountNet = %d, want 95000 (seller fee flat 5%%)", tx.AmountNet)
	}
	if tx.ShippingFee != 0 {
		t.Errorf("ShippingFee must be 0 for services, got %d", tx.ShippingFee)
	}
}

func TestLockFundsServices_InvalidAmount(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, nil), &mockTransactionRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if _, err := u.LockFundsServices(ctx, "b", "s", 0, true, "BRONZE", "REKBERPAY", "idem", validServicesDetail(), validServicesMilestones(0)); err == nil {
		t.Fatal("expected error for amount = 0")
	}
}

func TestLockFundsServices_CreateError(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onCreateTransaction: func(ctx context.Context, tx *domain.Transaction) error {
			return errors.New("concurrent insert conflict")
		},
	}
	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if _, err := u.LockFundsServices(ctx, "b", "s", 50000, true, "BRONZE", "REKBERPAY", "idem", validServicesDetail(), validServicesMilestones(50000)); err == nil {
		t.Fatal("expected error when CreateTransaction fails")
	}
}

func TestReleaseMilestoneFunds_Success(t *testing.T) {
	ctx := context.Background()
	milestoneID := "ms-uuid-1"

	txRepo := &mockTransactionRepo{
		onGetMilestoneByID: func(ctx context.Context, id string) (*domain.ServiceMilestone, error) {
			return &domain.ServiceMilestone{
				ID:             milestoneID,
				TransactionID:  "parent-tx-1",
				MilestoneIndex: 1,
				Title:          "UI Design",
				Amount:         30000000,
				Status:         "PENDING",
			}, nil
		},
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: id, SellerID: "seller-uuid"}, nil
		},
		onUpdateMilestoneStatus: func(ctx context.Context, id string, status string) error {
			if status != "RELEASED" {
				t.Errorf("milestone status = %s, want RELEASED", status)
			}
			return nil
		},
	}

	var capturedAmount int64
	var capturedModifier int64
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			capturedAmount = rec.Amount
			capturedModifier = modifier
			return nil
		},
	}

	var capturedRevenue int64
	financeRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error {
			capturedRevenue = revenueDelta
			if escrowDelta != -30000000 {
				t.Errorf("escrow delta = %d, want -30000000", escrowDelta)
			}
			return nil
		},
	}

	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(txRepo, walletRepo, financeRepo, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ReleaseMilestoneFunds(ctx, milestoneID); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	// Milestone 30.000.000; flat 5% commission = 1.500.000 -> seller nets 28.500.000.
	if capturedAmount != 28500000 {
		t.Errorf("amount wallet log = %d, want 28500000 (net after 5%% commission)", capturedAmount)
	}
	if capturedModifier != 28500000 {
		t.Errorf("modifier credit = %d, want 28500000 (net)", capturedModifier)
	}
	if capturedRevenue != 1500000 {
		t.Errorf("commission revenue = %d, want 1500000 (5%% of 30000000)", capturedRevenue)
	}
}

func TestReleaseMilestoneFunds_AlreadyReleased(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetMilestoneByID: func(ctx context.Context, id string) (*domain.ServiceMilestone, error) {
			return &domain.ServiceMilestone{ID: id, Status: "RELEASED"}, nil
		},
	}
	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ReleaseMilestoneFunds(ctx, "ms-done"); err == nil {
		t.Fatal("expected error because milestone already RELEASED")
	}
}

func TestConfirmPaymentServices_Success(t *testing.T) {
	ctx := context.Background()
	orderID := "REKBERKUY-SERVICE-abc12345"

	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:              "tx-uuid-svc-abc",
				BuyerID:         "buyer-uuid",
				MidtransOrderID: orderID,
				Status:          domain.StatusWaitingPayment,
				AmountGross:     100000,
				ServiceFee:      0,
			}, nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			if status != domain.StatusFundsLocked {
				t.Errorf("expected status transition to FUNDS_LOCKED, got: %s", status)
			}
			return nil
		},
	}

	var debitModifier int64
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			debitModifier = modifier
			return nil
		},
	}

	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(txRepo, walletRepo, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ConfirmPaymentServices(ctx, orderID); err != nil {
		t.Fatalf("expected payment confirmed successfully, got error: %v", err)
	}
	if debitModifier != -100000 {
		t.Errorf("buyer debit modifier = %d, want -100000", debitModifier)
	}
}

func TestConfirmPaymentServices_AlreadyProcessedIsIdempotent(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			// Midtrans retried the webhook after the payment was already locked.
			return &domain.Transaction{ID: "tx-already-locked", MidtransOrderID: oid, Status: domain.StatusFundsLocked}, nil
		},
	}
	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	// Idempotent: already past WAITING_PAYMENT -> no error (no-op).
	if err := u.ConfirmPaymentServices(ctx, "REKBERKUY-SERVICE-xyz"); err != nil {
		t.Fatalf("expected no-op nil (idempotent replay), got error: %v", err)
	}
}
