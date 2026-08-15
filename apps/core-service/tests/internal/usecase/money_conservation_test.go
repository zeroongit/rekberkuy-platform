package usecase_test

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// MONEY-CONSERVATION PROPERTY TESTS
// ---------------------------------------------------------------------------
// Invariant: for any closed transaction lifecycle, money is conserved —
//     Σ(wallet deltas) + platform revenue + net escrow change == 0
// i.e. no rupiah is created or destroyed. This catches formula errors in the
// fee/escrow booking that a hardcoded-value unit test can miss (a wrong
// expected value baked into a unit test would pass; the conservation law
// cannot). MidtransFee is held at 0 so the invariant is clean.
//
// (Fix A — dispute over-refund — is NOT covered here: over-refunding still
// conserves money by driving escrow negative. It is covered by the dedicated
// TestDispute_ResolveDispute_RefundBuyer_MinusReleasedMilestones instead.)
// ============================================================================

// accumulator wallets
type conservationWalletRepo struct {
	domain.WalletRepository
	deltas map[string]int64
}

func (m *conservationWalletRepo) UpdateBalanceTx(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
	m.deltas[rec.WalletID] += modifier
	return nil
}

// MarkVendorAllocationClaimed is invoked when a vendor payout settles — the
// pledge draw-down moves no money, so the conservation ledger is unaffected.
func (m *conservationWalletRepo) MarkVendorAllocationClaimed(ctx context.Context, transactionID, vendorID string, amount int64) error {
	return nil
}

// GetVendorAllocationsByTxID feeds the invoice cap; no pledges in this scenario.
func (m *conservationWalletRepo) GetVendorAllocationsByTxID(ctx context.Context, transactionID string) ([]*domain.EventVendorAllocation, error) {
	return nil, nil
}

// accumulator finance
type conservationFinanceRepo struct {
	domain.FinanceRepository
	escrow, revenue int64
}

func (m *conservationFinanceRepo) UpdatePlatformFinance(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error {
	m.escrow += escrowDelta
	m.revenue += revenueDelta
	return nil
}

func walletSum(m map[string]int64) int64 {
	var s int64
	for _, d := range m {
		s += d
	}
	return s
}

// Goods: WAITING_PAYMENT -> FUNDS_LOCKED -> RELEASED must conserve money.
// Catches Fix B (revenue must equal the full retention AmountGross-AmountNet;
// if it stayed at the buyer-protection fee only, the sum would be non-zero).
func TestMoneyConservation_GoodsLifecycle(t *testing.T) {
	ctx := context.Background()
	const amountGross, amountNet = 112500, 97500

	wallet := &conservationWalletRepo{deltas: map[string]int64{}}
	fin := &conservationFinanceRepo{}
	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-g", BuyerID: "buyer", Status: domain.StatusWaitingPayment, AmountGross: amountGross, ServiceFee: 2500}, nil
		},
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-g", BuyerID: "buyer", SellerID: "seller", Status: domain.StatusFundsLocked, AmountGross: amountGross, AmountNet: amountNet}, nil
		},
	}
	u := usecase.NewTransactionGoodsUsecase(newMockUnitOfWork(txRepo, wallet, fin, nil), txRepo, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ConfirmPaymentGoods(ctx, "order-g"); err != nil {
		t.Fatalf("lock failed: %v", err)
	}
	if err := u.ReleaseFundsGoods(ctx, "tx-g"); err != nil {
		t.Fatalf("release failed: %v", err)
	}

	got := walletSum(wallet.deltas) + fin.revenue + fin.escrow
	if got != 0 {
		t.Errorf("goods money NOT conserved: wallets=%d revenue=%d escrow=%d -> sum=%d (want 0)",
			walletSum(wallet.deltas), fin.revenue, fin.escrow, got)
	}
}

// Events: WAITING_PAYMENT -> FUNDS_LOCKED -> vendor payout must conserve money.
// Catches Fix C (escrow must release vendorTotal + 5% fee, and the 5% must be
// booked as revenue; the surplus stays held in escrow. If the full AmountGross
// left escrow with 0 revenue — the old bug — the sum would be non-zero).
func TestMoneyConservation_EventLifecycle(t *testing.T) {
	ctx := context.Background()
	const amountGross = 50000000

	wallet := &conservationWalletRepo{deltas: map[string]int64{}}
	fin := &conservationFinanceRepo{}
	txRepo := &mockTransactionRepo{
		onGetByMidtransOrderID: func(ctx context.Context, oid string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-e", BuyerID: "eo", Status: domain.StatusWaitingPayment, AmountGross: amountGross}, nil
		},
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-e", BuyerID: "eo", Status: domain.StatusFundsLocked, AmountGross: amountGross}, nil
		},
		onGetEventVendorPayouts: func(ctx context.Context, id string) ([]domain.EventVendorPayout, error) {
			v1, v2 := "vendor-1", "vendor-2"
			return []domain.EventVendorPayout{
				{ID: "p1", VendorUserID: &v1, AmountRequested: 5000000, Status: "PENDING"},
				{ID: "p2", VendorUserID: &v2, AmountRequested: 3000000, Status: "PENDING"},
			}, nil
		},
	}
	u := usecase.NewTransactionEventsUsecase(newMockUnitOfWork(txRepo, wallet, fin, nil), txRepo, wallet, usecase.NewFinanceCalculator(), &mockFraudClient{}, &mockRelayer{})

	if err := u.ConfirmPaymentEvents(ctx, "order-e"); err != nil {
		t.Fatalf("lock failed: %v", err)
	}
	if err := u.ProcessEventVendorPayouts(ctx, "tx-e"); err != nil {
		t.Fatalf("payout failed: %v", err)
	}

	got := walletSum(wallet.deltas) + fin.revenue + fin.escrow
	if got != 0 {
		t.Errorf("event money NOT conserved: wallets=%d revenue=%d escrow=%d -> sum=%d (want 0)",
			walletSum(wallet.deltas), fin.revenue, fin.escrow, got)
	}
	// Escrow must hold the surplus (not zero): 50M - 8M vendors - 2.5M fee = 39.5M.
	if fin.escrow != 39500000 {
		t.Errorf("event escrow residue = %d, want 39500000 (surplus held for EO/participant distribution)", fin.escrow)
	}
}
