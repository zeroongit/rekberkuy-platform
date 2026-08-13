package usecase_test

import (
	"context"
	"errors"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// TRANSACTION GOODS USECASE — UNIT TESTS
// ============================================================================
// Mock repositories moved to mocks_test.go (shared across all test files
// in the usecase_test package).

func TestLockFundsGoods_Success(t *testing.T) {
	ctx := context.Background()

	txRepo := &mockTransactionRepo{}
	walletRepo := &mockWalletRepo{}
	financeRepo := &mockFinanceRepo{}
	calc := usecase.NewFinanceCalculator()

	uow := newMockUnitOfWork(txRepo, walletRepo, financeRepo, nil)
	u := usecase.NewTransactionGoodsUsecase(uow, txRepo, calc, &mockFraudClient{}, &mockRelayer{})

	tx, err := u.LockFundsGoods(
		ctx,
		"buyer-uuid-123",
		"seller-uuid-456",
		100000,
		true,
		"BRONZE",
		10000,
		"REKBERPAY",
		"idem-key-goods-01",
	)

	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if tx.Status != domain.StatusWaitingPayment {
		t.Errorf("expected status %s, got: %s", domain.StatusWaitingPayment, tx.Status)
	}

	if tx.Type != domain.TypeGoods {
		t.Errorf("expected type %s, got: %s", domain.TypeGoods, tx.Type)
	}

	// BRONZE + RekberPay -> flat buyer fee Rp2,500
	// amountGross = amountBase(100000) + buyerFee(2500) + shippingFee(10000) = 112500
	if tx.AmountGross != 112500 {
		t.Errorf("AmountGross wrong: expected 112500, got %d", tx.AmountGross)
	}

	// seller fee BRONZE = 10% of amountBase = 10000; amountNet = 100000 - 10000 = 90000
	if tx.AmountNet != 90000 {
		t.Errorf("AmountNet wrong: expected 90000, got %d", tx.AmountNet)
	}

	if tx.MidtransOrderID == "" {
		t.Error("MidtransOrderID must not be empty")
	}
}

func TestLockFundsGoods_InvalidAmount(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	_, err := u.LockFundsGoods(ctx, "b-id", "s-id", 0, true, "BRONZE", 10000, "REKBERPAY", "idem-02")
	if err == nil {
		t.Fatal("expected error because goods transaction amount = 0, but got nil")
	}
}

func TestLockFundsGoods_CreateTransactionError(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onCreateTransaction: func(ctx context.Context, tx *domain.Transaction) error {
			return errors.New("db down")
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	_, err := u.LockFundsGoods(ctx, "b-id", "s-id", 50000, true, "BRONZE", 0, "REKBERPAY", "idem-err")
	if err == nil {
		t.Fatal("expected error when CreateTransaction fails")
	}
}

func TestConfirmPaymentGoods_Success(t *testing.T) {
	ctx := context.Background()
	orderID := "REKBERKUY-GOODS-abc12345"

	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:              "tx-uuid-goods-abc",
				BuyerID:         "buyer-uuid",
				MidtransOrderID: orderID,
				Status:          domain.StatusWaitingPayment,
				AmountGross:     112500,
				ServiceFee:      2500,
			}, nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			if status != domain.StatusFundsLocked {
				t.Errorf("expected status transition to FUNDS_LOCKED, got: %s", status)
			}
			return nil
		},
	}

	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ConfirmPaymentGoods(ctx, orderID); err != nil {
		t.Fatalf("expected payment confirmed successfully, got error: %v", err)
	}
}

func TestConfirmPaymentGoods_AlreadyProcessedIsIdempotent(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			// Midtrans retried the webhook after the payment was already locked.
			return &domain.Transaction{ID: "tx-already-locked", MidtransOrderID: oid, Status: domain.StatusFundsLocked}, nil
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	// Idempotent: already past WAITING_PAYMENT -> no error (no-op), so the webhook
	// handler returns 200 and Midtrans stops retrying.
	if err := u.ConfirmPaymentGoods(ctx, "REKBERKUY-GOODS-xyz"); err != nil {
		t.Fatalf("expected no-op nil (idempotent replay), got error: %v", err)
	}
}

func TestReleaseFundsGoods_Success(t *testing.T) {
	ctx := context.Background()
	txID := "tx-uuid-goods-release"

	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:          txID,
				SellerID:    "seller-uuid",
				Status:      domain.StatusFundsLocked,
				AmountGross: 112500,
				AmountNet:   97500,
				ServiceFee:  2500,
				MidtransFee: 0,
			}, nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			if status != domain.StatusReleased {
				t.Errorf("expected status transition to RELEASED, got: %s", status)
			}
			return nil
		},
	}

	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ReleaseFundsGoods(ctx, txID); err != nil {
		t.Fatalf("expected fund release succeeded, got error: %v", err)
	}
}

func TestReleaseFundsGoods_WrongStatus(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: id, Status: domain.StatusWaitingPayment}, nil
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	err := u.ReleaseFundsGoods(ctx, "tx-not-locked")
	if err == nil {
		t.Fatal("expected error because status is not FUNDS_LOCKED")
	}
}

func TestReleaseFundsGoods_FraudRefusalBlocksRelease(t *testing.T) {
	ctx := context.Background()
	walletDebited := false
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: id, Status: domain.StatusFundsLocked, SellerID: "seller-1", BuyerID: "buyer-1", AmountGross: 100000, AmountNet: 90000}, nil
		},
	}
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			walletDebited = true
			return nil
		},
	}
	fraud := &mockFraudClient{
		onAnalyze: func(ctx context.Context, userID string, amount int64) (float64, bool, error) {
			return 0.95, false, nil // suspicious transaction
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, walletRepo, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), fraud, &mockRelayer{})

	err := u.ReleaseFundsGoods(ctx, "tx-fraud")
	if err == nil {
		t.Fatal("expected error because fraud scoring rejected release")
	}
	if walletDebited {
		t.Error("seller wallet must not be credited when release is rejected by fraud")
	}
}

// ---------------------------------------------------------------------------
// Member (USER-role) selling cap — MaxMemberEventLimit (Rp 10.000.000).
// A plain USER may sell up to the cap; verified sellers are unrestricted.
// The cap is enforced identically for goods, services, and events (shared
// helper), so we exercise it through the goods lock flow.
// ---------------------------------------------------------------------------

func TestLockFundsGoods_UserSellerOverCapRejected(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		onGetProfileByID: func(ctx context.Context, id string) (*domain.UserProfile, error) {
			return &domain.UserProfile{ID: id, Role: domain.RoleUser}, nil
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, userRepo), &mockTransactionRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	// 15M > 10M cap, seller is a regular USER -> must be rejected.
	if _, err := u.LockFundsGoods(ctx, "buyer-1", "user-seller", 15000000, true, "BRONZE", 0, "REKBERPAY", "idem-cap-1"); err == nil {
		t.Fatal("expected error: regular USER seller over the 10M cap must be rejected")
	}
}

func TestLockFundsGoods_UserSellerAtCapAllowed(t *testing.T) {
	ctx := context.Background()
	// At the cap exactly (10M) the fast path skips the profile lookup, so a nil
	// user repo is fine — and proves the boundary is inclusive.
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, nil), &mockTransactionRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if _, err := u.LockFundsGoods(ctx, "buyer-1", "user-seller", domain.MaxMemberEventLimit, true, "BRONZE", 0, "REKBERPAY", "idem-cap-2"); err != nil {
		t.Fatalf("expected success at exactly the 10M cap, got: %v", err)
	}
}

func TestLockFundsGoods_VerifiedSellerOverCapAllowed(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		onGetProfileByID: func(ctx context.Context, id string) (*domain.UserProfile, error) {
			return &domain.UserProfile{ID: id, Role: domain.RoleVerifiedMerchant}, nil
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, userRepo), &mockTransactionRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	// 15M over the cap, but the seller is a Verified Merchant -> allowed.
	if _, err := u.LockFundsGoods(ctx, "buyer-1", "verified-seller", 15000000, true, "GOLD", 0, "REKBERPAY", "idem-cap-3"); err != nil {
		t.Fatalf("expected verified merchant to be allowed over the cap, got: %v", err)
	}
}
