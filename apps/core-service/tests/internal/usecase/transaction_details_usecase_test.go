package usecase_test

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// DETAIL CONTAINERS + VENDOR INVOICE — UNIT TESTS
// ============================================================================

func TestLockFundsGoods_MissingDetail(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if _, err := u.LockFundsGoods(ctx, "b", "s", 100000, true, "BRONZE", 0, "REKBERPAY", "idem", nil); err == nil {
		t.Fatal("expected error when goods detail is missing")
	}

	incomplete := validGoodsDetail()
	incomplete.ShippingAddress = ""
	if _, err := u.LockFundsGoods(ctx, "b", "s", 100000, true, "BRONZE", 0, "REKBERPAY", "idem", incomplete); err == nil {
		t.Fatal("expected error when shipping address is empty")
	}
}

func TestLockFundsGoods_DetailLinkedToMaster(t *testing.T) {
	ctx := context.Background()

	var capturedTG *domain.TransactionGoods
	txRepo := &mockTransactionRepo{
		onCreateGoodsDetail: func(ctx context.Context, tg *domain.TransactionGoods) error {
			capturedTG = tg
			return nil
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	tx, err := u.LockFundsGoods(ctx, "b", "s", 100000, true, "BRONZE", 0, "REKBERPAY", "idem", validGoodsDetail())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if capturedTG == nil {
		t.Fatal("CreateGoodsDetail was not called")
	}
	if capturedTG.TransactionID != tx.ID {
		t.Errorf("detail TransactionID = %s, want master ID %s", capturedTG.TransactionID, tx.ID)
	}
}

func TestLockFundsServices_MilestoneSumMismatch(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, nil), &mockTransactionRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	// milestones sum to 90000 but base amount is 100000 -> must be rejected
	ms := []domain.ServiceMilestone{{Title: "A", Amount: 40000}, {Title: "B", Amount: 50000}}
	if _, err := u.LockFundsServices(ctx, "b", "s", 100000, true, "BRONZE", "REKBERPAY", "idem", validServicesDetail(), ms); err == nil {
		t.Fatal("expected error when milestone total != amount base")
	}

	if _, err := u.LockFundsServices(ctx, "b", "s", 100000, true, "BRONZE", "REKBERPAY", "idem", validServicesDetail(), nil); err == nil {
		t.Fatal("expected error when milestones list is empty")
	}
}

func TestLockFundsServices_MilestonesNormalized(t *testing.T) {
	ctx := context.Background()

	var captured []domain.ServiceMilestone
	txRepo := &mockTransactionRepo{
		onCreateServicesDetail: func(ctx context.Context, ts *domain.TransactionServices, milestones []domain.ServiceMilestone) error {
			captured = milestones
			return nil
		},
	}
	u := usecase.NewTransactionServicesUsecase(newMockUnitOfWork(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	tx, err := u.LockFundsServices(ctx, "b", "s", 100000, true, "BRONZE", "REKBERPAY", "idem", validServicesDetail(), validServicesMilestones(100000))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(captured) != 2 {
		t.Fatalf("captured %d milestones, want 2", len(captured))
	}
	for i, m := range captured {
		if m.MilestoneIndex != i+1 {
			t.Errorf("milestone %d index = %d, want %d", i, m.MilestoneIndex, i+1)
		}
		if m.Status != "PENDING" {
			t.Errorf("milestone %d status = %s, want PENDING", i, m.Status)
		}
		if m.ID == "" || m.TransactionID != tx.ID {
			t.Errorf("milestone %d must receive an ID and the master transaction ID", i)
		}
	}
}

func TestLockFundsEvents_AllocationExceedsBase(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, nil), &mockTransactionRepo{}, &mockWalletRepo{}, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	allocs := []domain.EventVendorAllocation{
		{VendorID: "v-1", AllocatedAmount: 60000},
		{VendorID: "v-2", AllocatedAmount: 60000},
	}
	if _, err := u.LockFundsEvents(ctx, "b", "s", 100000, true, "GOLD", "REKBERPAY", "idem", validEventsDetail(), allocs); err == nil {
		t.Fatal("expected error when allocation total exceeds amount base")
	}
}

// --- Vendor invoice submission ------------------------------------------------

func newEventsUsecaseForInvoice(txRepo *mockTransactionRepo, walletRepo *mockWalletRepo) *usecase.TransactionEventsUsecase {
	uow := newMockUnitOfWork(txRepo, walletRepo, &mockFinanceRepo{}, nil)
	return usecase.NewTransactionEventsUsecase(uow, txRepo, walletRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})
}

func validEventTxForInvoice() *domain.Transaction {
	return &domain.Transaction{
		ID: "evt-tx-1", BuyerID: "buyer-1", SellerID: "eo-1",
		Type: domain.TypeEvents, Status: domain.StatusFundsLocked,
		AmountBase: 1000000, AmountGross: 1000000, AmountNet: 1000000,
	}
}

func validVendorInvoice() *domain.EventVendorPayout {
	return &domain.EventVendorPayout{
		TransactionID:       "evt-tx-1",
		VendorName:          "Katering Sedap",
		VendorBankName:      "BCA",
		VendorAccountNumber: "1234567890",
		AmountRequested:     300000,
		ExpenseDescription:  "Catering for 100 pax",
		InvoiceFileURL:      "https://gateway.pinata.cloud/ipfs/QmExample",
	}
}

func TestSubmitEventVendorInvoice_Success(t *testing.T) {
	ctx := context.Background()

	var created *domain.EventVendorPayout
	txRepo := &mockTransactionRepo{
		onGetTransactionByID:    func(ctx context.Context, id string) (*domain.Transaction, error) { return validEventTxForInvoice(), nil },
		onGetActivePayoutsTotal: func(ctx context.Context, txID string) (int64, error) { return 0, nil },
	}
	walletRepo := &mockWalletRepo{
		onCreateVendorPayoutRecord: func(ctx context.Context, p *domain.EventVendorPayout) error {
			created = p
			return nil
		},
	}
	u := newEventsUsecaseForInvoice(txRepo, walletRepo)

	if err := u.SubmitEventVendorInvoice(ctx, "eo-1", validVendorInvoice()); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if created == nil {
		t.Fatal("CreateVendorPayoutRecord was not called")
	}
	if created.Status != domain.VendorPayoutPending {
		t.Errorf("payout status = %s, want PENDING", created.Status)
	}
	if created.PayoutPhase != "FINAL_SETTLEMENT" {
		t.Errorf("payout phase = %s, want FINAL_SETTLEMENT default", created.PayoutPhase)
	}
	if created.ID == "" {
		t.Error("payout ID must be generated")
	}
}

func TestSubmitEventVendorInvoice_OnlyOrganizer(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) { return validEventTxForInvoice(), nil },
	}
	u := newEventsUsecaseForInvoice(txRepo, &mockWalletRepo{})

	if err := u.SubmitEventVendorInvoice(ctx, "someone-else", validVendorInvoice()); err == nil {
		t.Fatal("expected error when requester is not the event organizer")
	}
}

func TestSubmitEventVendorInvoice_ExceedsPayableEscrow(t *testing.T) {
	ctx := context.Background()
	// gross 1,000,000 -> 5% fee = 50,000 -> payable cap 950,000.
	// Existing active claims already 800,000; a new 300,000 must be rejected.
	txRepo := &mockTransactionRepo{
		onGetTransactionByID:    func(ctx context.Context, id string) (*domain.Transaction, error) { return validEventTxForInvoice(), nil },
		onGetActivePayoutsTotal: func(ctx context.Context, txID string) (int64, error) { return 800000, nil },
	}
	u := newEventsUsecaseForInvoice(txRepo, &mockWalletRepo{})

	if err := u.SubmitEventVendorInvoice(ctx, "eo-1", validVendorInvoice()); err == nil {
		t.Fatal("expected error when total claims exceed the payable escrow")
	}
}

func TestSubmitEventVendorInvoice_RequiresFundsLocked(t *testing.T) {
	ctx := context.Background()
	waiting := validEventTxForInvoice()
	waiting.Status = domain.StatusWaitingPayment
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) { return waiting, nil },
	}
	u := newEventsUsecaseForInvoice(txRepo, &mockWalletRepo{})

	if err := u.SubmitEventVendorInvoice(ctx, "eo-1", validVendorInvoice()); err == nil {
		t.Fatal("expected error when escrow is not FUNDS_LOCKED")
	}
}

func TestSubmitEventVendorInvoice_OnlyEventTransactions(t *testing.T) {
	ctx := context.Background()
	goods := validEventTxForInvoice()
	goods.Type = domain.TypeGoods
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) { return goods, nil },
	}
	u := newEventsUsecaseForInvoice(txRepo, &mockWalletRepo{})

	if err := u.SubmitEventVendorInvoice(ctx, "eo-1", validVendorInvoice()); err == nil {
		t.Fatal("expected error when the transaction is not an event escrow")
	}
}

// --- Pledge-aware escrow cap --------------------------------------------------

func TestSubmitEventVendorInvoice_OtherVendorsPledgesReserveEscrow(t *testing.T) {
	ctx := context.Background()
	// gross 1,000,000 -> payable cap 950,000. Another vendor's unsettled
	// pledge holds 700,000, leaving 250,000 for new invoices — a 300,000
	// invoice from a different vendor must be rejected.
	txRepo := &mockTransactionRepo{
		onGetTransactionByID:    func(ctx context.Context, id string) (*domain.Transaction, error) { return validEventTxForInvoice(), nil },
		onGetActivePayoutsTotal: func(ctx context.Context, txID string) (int64, error) { return 0, nil },
	}
	walletRepo := &mockWalletRepo{
		onGetVendorAllocations: func(ctx context.Context, txID string) ([]*domain.EventVendorAllocation, error) {
			return []*domain.EventVendorAllocation{
				{VendorID: "other-vendor", AllocatedAmount: 700000, Status: domain.VendorAllocationPledged},
			}, nil
		},
	}
	u := newEventsUsecaseForInvoice(txRepo, walletRepo)

	inv := validVendorInvoice() // 300,000 from a vendor with no pledge
	vendorID := "vendor-without-pledge"
	inv.VendorUserID = &vendorID
	if err := u.SubmitEventVendorInvoice(ctx, "eo-1", inv); err == nil {
		t.Fatal("expected error: invoice + other vendors' pledges exceed the payable escrow")
	}
}

func TestSubmitEventVendorInvoice_OwnPledgeIsNetted(t *testing.T) {
	ctx := context.Background()
	// The submitting vendor's own pledge must NOT double-count: a 300,000
	// invoice from a vendor holding a 300,000 pledge fits even though
	// pledges + invoices nominally equal 600,000 of a 950,000 cap.
	txRepo := &mockTransactionRepo{
		onGetTransactionByID:    func(ctx context.Context, id string) (*domain.Transaction, error) { return validEventTxForInvoice(), nil },
		onGetActivePayoutsTotal: func(ctx context.Context, txID string) (int64, error) { return 0, nil },
	}
	vendorID := "vendor-a"
	walletRepo := &mockWalletRepo{
		onGetVendorAllocations: func(ctx context.Context, txID string) ([]*domain.EventVendorAllocation, error) {
			return []*domain.EventVendorAllocation{
				{VendorID: vendorID, AllocatedAmount: 300000, Status: domain.VendorAllocationPledged},
			}, nil
		},
		onCreateVendorPayoutRecord: func(ctx context.Context, p *domain.EventVendorPayout) error { return nil },
	}
	u := newEventsUsecaseForInvoice(txRepo, walletRepo)

	inv := validVendorInvoice() // 300,000
	inv.VendorUserID = &vendorID
	if err := u.SubmitEventVendorInvoice(ctx, "eo-1", inv); err != nil {
		t.Fatalf("own pledge must be netted, expected success, got: %v", err)
	}
}

func TestSubmitEventVendorInvoice_ClaimedPledgesDoNotReserve(t *testing.T) {
	ctx := context.Background()
	// A fully claimed pledge no longer holds escrow: only 100,000 of a
	// 400,000 pledge remains unsettled -> reservation is 100,000.
	txRepo := &mockTransactionRepo{
		onGetTransactionByID:    func(ctx context.Context, id string) (*domain.Transaction, error) { return validEventTxForInvoice(), nil },
		onGetActivePayoutsTotal: func(ctx context.Context, txID string) (int64, error) { return 0, nil },
	}
	walletRepo := &mockWalletRepo{
		onGetVendorAllocations: func(ctx context.Context, txID string) ([]*domain.EventVendorAllocation, error) {
			return []*domain.EventVendorAllocation{
				{VendorID: "other", AllocatedAmount: 400000, ActualPaidAmount: 300000, Status: domain.VendorAllocationPledged},
			}, nil
		},
		onCreateVendorPayoutRecord: func(ctx context.Context, p *domain.EventVendorPayout) error { return nil },
	}
	u := newEventsUsecaseForInvoice(txRepo, walletRepo)

	// 900,000 > cap 950,000 - 100,000 reservation = 850,000 -> must be rejected.
	inv := validVendorInvoice()
	inv.AmountRequested = 900000
	if err := u.SubmitEventVendorInvoice(ctx, "eo-1", inv); err == nil {
		t.Fatal("expected error: remaining pledge reservation must still cap invoices")
	}
}
