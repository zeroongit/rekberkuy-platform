package usecase_test

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// TRANSACTION EVENTS USECASE — UNIT TESTS
// ============================================================================

func TestLockFundsEvents_Success(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{}
	// 50M event exceeds the USER selling cap; the seller here is an Event
	// Organiser (not a plain USER) so the cap must not reject it.
	userRepo := &mockUserRepo{
		onGetProfileByID: func(ctx context.Context, id string) (*domain.UserProfile, error) {
			return &domain.UserProfile{ID: id, Role: domain.RoleEventOrganizer}, nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, userRepo), txRepo, &mockWalletRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	// Events: buyer fee = 0, seller fee = 0 -> gross = net = base
	tx, err := u.LockFundsEvents(ctx, "buyer-1", "seller-1", 50000000, true, "GOLD", "REKBERPAY", "idem-evt-1", validEventsDetail(), nil)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if tx.Type != domain.TypeEvents {
		t.Errorf("type = %s, want %s", tx.Type, domain.TypeEvents)
	}
	if tx.AmountGross != 50000000 {
		t.Errorf("AmountGross = %d, want 50000000", tx.AmountGross)
	}
	if tx.AmountNet != 50000000 {
		t.Errorf("AmountNet = %d, want 50000000 (events fee 0)", tx.AmountNet)
	}
}

func TestLockFundsEvents_InvalidAmount(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, nil), &mockTransactionRepo{}, &mockWalletRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if _, err := u.LockFundsEvents(ctx, "b", "s", -100, true, "GOLD", "REKBERPAY", "idem", validEventsDetail(), nil); err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestProcessEventVendorPayouts_Success(t *testing.T) {
	ctx := context.Background()
	txID := "evt-tx-1"

	releasedStatuses := map[string]string{}
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:          txID,
				Status:      domain.StatusFundsLocked,
				AmountGross: 50000000,
				ServiceFee:  0,
			}, nil
		},
		onGetEventVendorPayouts: func(ctx context.Context, id string) ([]domain.EventVendorPayout, error) {
			return []domain.EventVendorPayout{
				{ID: "pay-1", TransactionID: txID, VendorUserID: strPtr("vendor-user-1"), VendorName: "Katering A", AmountRequested: 5000000, Status: "PENDING"},
				{ID: "pay-2", TransactionID: txID, VendorUserID: strPtr("vendor-user-2"), VendorName: "Dekor B", AmountRequested: 3000000, Status: "PENDING"},
			}, nil
		},
		onUpdateEventVendorPayoutStatus: func(ctx context.Context, id string, status string) error {
			releasedStatuses[id] = status
			return nil
		}, onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			if status != domain.StatusReleased {
				t.Errorf("parent transaction status = %s, want RELEASED", status)
			}
			return nil
		},
	}

	creditedWallets := map[string]bool{}
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			creditedWallets[rec.WalletID] = true
			return nil
		},
	}

	var revenue, escrowDelta int64
	financeRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, e, r, m int64) error {
			escrowDelta, revenue = e, r
			return nil
		},
	}

	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, walletRepo, financeRepo, nil), txRepo, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ProcessEventVendorPayouts(ctx, txID); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(creditedWallets) != 2 {
		t.Errorf("expected 2 vendor wallets credited, got %d", len(creditedWallets))
	}
	if !creditedWallets["vendor-user-1"] || !creditedWallets["vendor-user-2"] {
		t.Errorf("correct vendor user ID wallets must be credited, got: %v", creditedWallets)
	}
	if releasedStatuses["pay-1"] != domain.VendorPayoutApproved || releasedStatuses["pay-2"] != domain.VendorPayoutApproved {
		t.Errorf("payout status not updated to APPROVED: %v", releasedStatuses)
	}
	// Vendor total = 5M + 3M = 8M; platform fee = 5% of 50M = 2.5M (revenue);
	// escrow releases vendor total + platform fee = 10.5M; the surplus stays held.
	if revenue != 2500000 {
		t.Errorf("platform fee revenue = %d, want 2500000 (5%% of 50M)", revenue)
	}
	if escrowDelta != -10500000 {
		t.Errorf("escrow delta = %d, want -10500000 (vendor total 8M + 2.5M fee)", escrowDelta)
	}
}

func TestProcessEventVendorPayouts_WrongStatus(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: id, Status: domain.StatusWaitingPayment}, nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, &mockWalletRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ProcessEventVendorPayouts(ctx, "tx-not-locked"); err == nil {
		t.Fatal("expected error because status is not FUNDS_LOCKED")
	}
}

func TestProcessEventVendorPayouts_NoPayouts(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: id, Status: domain.StatusFundsLocked}, nil
		},
		onGetEventVendorPayouts: func(ctx context.Context, id string) ([]domain.EventVendorPayout, error) {
			return nil, nil // no invoices/bills
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, &mockWalletRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ProcessEventVendorPayouts(ctx, "tx-empty"); err == nil {
		t.Fatal("expected error because no vendor bills")
	}
}

func TestProcessEventVendorPayouts_SkipsAlreadyApproved(t *testing.T) {
	ctx := context.Background()
	txID := "evt-tx-2"

	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: txID, Status: domain.StatusFundsLocked, AmountGross: 1000000}, nil
		},
		onGetEventVendorPayouts: func(ctx context.Context, id string) ([]domain.EventVendorPayout, error) {
			return []domain.EventVendorPayout{
				{ID: "pay-1", VendorName: "Sudah Bayar", AmountRequested: 500000, Status: "APPROVED"},
				{ID: "pay-2", VendorUserID: strPtr("vendor-user-x"), VendorName: "Belum Bayar", AmountRequested: 300000, Status: "PENDING"},
			}, nil
		},
	}

	creditCalls := 0
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			creditCalls++
			return nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, walletRepo, &mockFinanceRepo{}, nil), txRepo, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ProcessEventVendorPayouts(ctx, txID); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if creditCalls != 1 {
		t.Errorf("expected only 1 credit (skip APPROVED), got %d", creditCalls)
	}
}

func TestReleaseEventMilestonePayout_Success(t *testing.T) {
	ctx := context.Background()
	payoutID := "pay-single-1"

	txRepo := &mockTransactionRepo{}
	walletRepo := &mockWalletRepo{
		onGetVendorPayoutByID: func(ctx context.Context, id string) (*domain.EventVendorPayout, error) {
			return &domain.EventVendorPayout{
				ID:              payoutID,
				TransactionID:   "evt-tx-3",
				VendorUserID:    strPtr("vendor-user-3"),
				VendorName:      "Foto C",
				AmountRequested: 7500000,
				Status:          "PENDING",
				PayoutPhase:     "DP",
			}, nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, walletRepo, &mockFinanceRepo{}, nil), txRepo, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ReleaseEventMilestonePayout(ctx, payoutID); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestReleaseEventMilestonePayout_AlreadyApproved(t *testing.T) {
	ctx := context.Background()
	walletRepo := &mockWalletRepo{
		onGetVendorPayoutByID: func(ctx context.Context, id string) (*domain.EventVendorPayout, error) {
			return &domain.EventVendorPayout{ID: id, Status: "APPROVED"}, nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, &mockFinanceRepo{}, nil), &mockTransactionRepo{}, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ReleaseEventMilestonePayout(ctx, "pay-done"); err == nil {
		t.Fatal("expected error because payout already APPROVED")
	}
}

// ---------------------------------------------------------------------------
// External vendor (VendorUserID=nil) -> do not credit wallet, mark PENDING_DISBURSEMENT
// ---------------------------------------------------------------------------

func TestProcessEventVendorPayouts_ExternalVendor_PendingDisbursement(t *testing.T) {
	ctx := context.Background()
	txID := "evt-tx-ext"

	parentReleased := false
	statusUpdates := map[string]string{}
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: txID, Status: domain.StatusFundsLocked, AmountGross: 1000000}, nil
		},
		onGetEventVendorPayouts: func(ctx context.Context, id string) ([]domain.EventVendorPayout, error) {
			return []domain.EventVendorPayout{
				{ID: "ext-1", VendorName: "Vendor Luar", AmountRequested: 800000, Status: "PENDING"}, // VendorUserID empty -> external
			}, nil
		},
		onUpdateEventVendorPayoutStatus: func(ctx context.Context, id string, status string) error {
			statusUpdates[id] = status
			return nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			parentReleased = true
			return nil
		},
	}
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			t.Errorf("external vendor must not credit wallet")
			return nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, walletRepo, &mockFinanceRepo{}, nil), txRepo, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ProcessEventVendorPayouts(ctx, txID); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if statusUpdates["ext-1"] != domain.VendorPayoutPendingDisbursement {
		t.Errorf("external vendor status = %s, want PENDING_DISBURSEMENT", statusUpdates["ext-1"])
	}
	if parentReleased {
		t.Error("parent transaction must not be released while external vendor pending disbursement exists")
	}
}

func TestReleaseEventMilestonePayout_ExternalVendor_PendingDisbursement(t *testing.T) {
	ctx := context.Background()
	walletRepo := &mockWalletRepo{
		onGetVendorPayoutByID: func(ctx context.Context, id string) (*domain.EventVendorPayout, error) {
			return &domain.EventVendorPayout{ID: id, AmountRequested: 500000, Status: "PENDING"}, nil // external
		},
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			t.Errorf("external vendor must not credit wallet")
			return nil
		},
		onUpdateVendorPayoutStatus: func(ctx context.Context, payoutID string, status string) error {
			if status != domain.VendorPayoutPendingDisbursement {
				t.Errorf("status = %s, want PENDING_DISBURSEMENT", status)
			}
			return nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, &mockFinanceRepo{}, nil), &mockTransactionRepo{}, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ReleaseEventMilestonePayout(ctx, "ext-pay-1"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestConfirmPaymentEvents_Success(t *testing.T) {
	ctx := context.Background()
	orderID := "REKBERKUY-EVENT-abc12345"

	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:              "tx-uuid-evt-abc",
				BuyerID:         "eo-uuid",
				MidtransOrderID: orderID,
				Status:          domain.StatusWaitingPayment,
				AmountGross:     50000000,
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

	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, walletRepo, &mockFinanceRepo{}, nil), txRepo, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ConfirmPaymentEvents(ctx, orderID); err != nil {
		t.Fatalf("expected payment confirmed successfully, got error: %v", err)
	}
	if debitModifier != -50000000 {
		t.Errorf("buyer debit modifier = %d, want -50000000", debitModifier)
	}
}

func TestConfirmPaymentEvents_AlreadyProcessedIsIdempotent(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-evt-already-locked", MidtransOrderID: oid, Status: domain.StatusFundsLocked}, nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, &mockWalletRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ConfirmPaymentEvents(ctx, "REKBERKUY-EVENT-xyz"); err != nil {
		t.Fatalf("expected no-op nil (idempotent replay), got error: %v", err)
	}
}

// strPtr helper to create a *string literal (VendorUserID field).
func strPtr(s string) *string { return &s }
