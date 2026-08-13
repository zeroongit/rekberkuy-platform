package usecase_test

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// DISPUTE USECASE — UNIT TESTS (behaviour: status transitions + money movement)
// ============================================================================

func TestDispute_OpenDispute_Success(t *testing.T) {
	ctx := context.Background()
	var frozenStatus domain.TransactionStatus
	var created *domain.Dispute
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusFundsLocked}, nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			frozenStatus = status
			return nil
		},
	}
	disputeRepo := &mockDisputeRepo{
		onCreateDispute: func(ctx context.Context, d *domain.Dispute) error {
			created = d
			return nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, disputeRepo), disputeRepo)

	got, err := u.OpenDispute(ctx, "tx-1", "buyer-1", "item not as described", nil)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if frozenStatus != domain.StatusDisputed {
		t.Errorf("transaction status = %s, want DISPUTED", frozenStatus)
	}
	if created == nil || created.Status != domain.DisputeStatusOpen {
		t.Errorf("dispute not created as OPEN: %+v", created)
	}
	if got == nil || got.RaisedBy != "buyer-1" || got.TargetPartyID == nil || *got.TargetPartyID != "seller-1" {
		t.Errorf("dispute target should be the seller, got %+v", got)
	}
}

func TestDispute_OpenDispute_NotFundsLocked(t *testing.T) {
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusReleased}, nil
		},
	}
	disputeRepo := &mockDisputeRepo{
		onCreateDispute: func(ctx context.Context, d *domain.Dispute) error {
			t.Error("must not create a dispute on a non-FUNDS_LOCKED transaction")
			return nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, disputeRepo), disputeRepo)
	if _, err := u.OpenDispute(context.Background(), "tx-1", "buyer-1", "reason", nil); err == nil {
		t.Fatal("expected error because transaction is not FUNDS_LOCKED")
	}
}

func TestDispute_OpenDispute_NotAParty(t *testing.T) {
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusFundsLocked}, nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(txRepo, &mockWalletRepo{}, &mockFinanceRepo{}, &mockDisputeRepo{}), &mockDisputeRepo{})
	if _, err := u.OpenDispute(context.Background(), "tx-1", "intruder", "reason", nil); err == nil {
		t.Fatal("expected error because raiser is neither buyer nor seller")
	}
}

func TestDispute_AcknowledgeDispute_Success(t *testing.T) {
	acknowledged := false
	disputeRepo := &mockDisputeRepo{
		onGetDisputeByID: func(ctx context.Context, id string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: "d-1", TransactionID: "tx-1", Status: domain.DisputeStatusOpen}, nil
		},
		onAcknowledgeDispute: func(ctx context.Context, id string) error {
			acknowledged = true
			return nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, disputeRepo), disputeRepo)
	if err := u.AcknowledgeDispute(context.Background(), "d-1"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !acknowledged {
		t.Error("dispute was not acknowledged")
	}
}

func TestDispute_AcknowledgeDispute_NotOpen(t *testing.T) {
	disputeRepo := &mockDisputeRepo{
		onGetDisputeByID: func(ctx context.Context, id string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: "d-1", Status: domain.DisputeStatusResolved}, nil
		},
		onAcknowledgeDispute: func(ctx context.Context, id string) error {
			t.Error("must not acknowledge a non-OPEN dispute")
			return nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, disputeRepo), disputeRepo)
	if err := u.AcknowledgeDispute(context.Background(), "d-1"); err == nil {
		t.Fatal("expected error because dispute is not OPEN")
	}
}

func TestDispute_ResolveDispute_RefundBuyer(t *testing.T) {
	ctx := context.Background()
	var creditedWallet string
	var creditedAmount int64
	var txStatus domain.TransactionStatus
	var resolved bool

	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusDisputed, AmountGross: 200000, AmountBase: 190000, ServiceFee: 10000, AmountNet: 180000}, nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			txStatus = status
			return nil
		},
	}
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			creditedWallet = rec.WalletID
			creditedAmount = modifier
			if rec.Type != domain.TxRefund {
				t.Errorf("ledger type = %s, want REFUND", rec.Type)
			}
			return nil
		},
	}
	disputeRepo := &mockDisputeRepo{
		onGetDisputeByID: func(ctx context.Context, id string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: "d-1", TransactionID: "tx-1", Status: domain.DisputeStatusUnderReview}, nil
		},
		onResolveDispute: func(ctx context.Context, id, adminID string, outcome domain.DisputeOutcome, summary string) error {
			resolved = true
			if outcome != domain.OutcomeRefundBuyer {
				t.Errorf("outcome = %s, want REFUND_BUYER", outcome)
			}
			return nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(txRepo, walletRepo, &mockFinanceRepo{}, disputeRepo), disputeRepo)

	if err := u.ResolveDispute(ctx, "d-1", "admin-1", domain.OutcomeRefundBuyer, "buyer was right"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if creditedWallet != "buyer-1" || creditedAmount != 200000 {
		t.Errorf("refund = %s/%d, want buyer-1/200000 (full gross — goods has no released milestones)", creditedWallet, creditedAmount)
	}
	if txStatus != domain.StatusRefunded {
		t.Errorf("transaction status = %s, want REFUNDED", txStatus)
	}
	if !resolved {
		t.Error("dispute was not resolved")
	}
}

// Regression: a milestone-based services transaction that already released some
// milestones must refund only the escrow still held, not the full AmountGross —
// otherwise the platform over-pays the buyer and escrow goes negative.
func TestDispute_ResolveDispute_RefundBuyer_MinusReleasedMilestones(t *testing.T) {
	ctx := context.Background()
	var creditedAmount int64
	var escrowDelta int64

	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-svc", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusDisputed, AmountGross: 10000000}, nil
		},
		onGetReleasedMilestonesTotal: func(ctx context.Context, txID string) (int64, error) {
			return 6000000, nil // Rp 6M already paid to the freelancer via released milestones
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error { return nil },
	}
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			creditedAmount = modifier
			return nil
		},
	}
	financeRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, escrowDelta_, revenueDelta, midtransFeeDelta int64) error {
			escrowDelta = escrowDelta_
			return nil
		},
	}
	disputeRepo := &mockDisputeRepo{
		onGetDisputeByID: func(ctx context.Context, id string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: "d-1", TransactionID: "tx-svc", Status: domain.DisputeStatusUnderReview}, nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(txRepo, walletRepo, financeRepo, disputeRepo), disputeRepo)

	if err := u.ResolveDispute(ctx, "d-1", "admin-1", domain.OutcomeRefundBuyer, "x"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if creditedAmount != 4000000 {
		t.Errorf("refund = %d, want 4000000 (10M gross - 6M already released)", creditedAmount)
	}
	if escrowDelta != -4000000 {
		t.Errorf("escrow delta = %d, want -4000000", escrowDelta)
	}
}

func TestDispute_ResolveDispute_ReleaseToSeller(t *testing.T) {
	ctx := context.Background()
	var creditedWallet string
	var creditedAmount int64
	var txStatus domain.TransactionStatus
	var revenue int64

	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: "tx-1", BuyerID: "buyer-1", SellerID: "seller-1", Status: domain.StatusDisputed, AmountGross: 200000, AmountNet: 180000, ServiceFee: 20000}, nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			txStatus = status
			return nil
		},
	}
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			creditedWallet = rec.WalletID
			creditedAmount = modifier
			if rec.Type != domain.TxReceiveFunds {
				t.Errorf("ledger type = %s, want RECEIVE_FUNDS", rec.Type)
			}
			return nil
		},
	}
	financeRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error {
			revenue = revenueDelta
			return nil
		},
	}
	disputeRepo := &mockDisputeRepo{
		onGetDisputeByID: func(ctx context.Context, id string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: "d-1", TransactionID: "tx-1", Status: domain.DisputeStatusUnderReview}, nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(txRepo, walletRepo, financeRepo, disputeRepo), disputeRepo)

	if err := u.ResolveDispute(ctx, "d-1", "admin-1", domain.OutcomeReleaseToSeller, "seller was right"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if creditedWallet != "seller-1" || creditedAmount != 180000 {
		t.Errorf("payout = %s/%d, want seller-1/180000 (AmountNet)", creditedWallet, creditedAmount)
	}
	if revenue != 20000 {
		t.Errorf("revenue = %d, want 20000 (full retention = AmountGross - AmountNet)", revenue)
	}
	if txStatus != domain.StatusReleased {
		t.Errorf("transaction status = %s, want RELEASED", txStatus)
	}
}

func TestDispute_ResolveDispute_NotUnderReview(t *testing.T) {
	disputeRepo := &mockDisputeRepo{
		onGetDisputeByID: func(ctx context.Context, id string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: "d-1", TransactionID: "tx-1", Status: domain.DisputeStatusOpen}, nil // not yet acknowledged
		},
	}
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, modifier int64) error {
			t.Error("must not move money before the dispute is UNDER_REVIEW")
			return nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(&mockTransactionRepo{}, walletRepo, &mockFinanceRepo{}, disputeRepo), disputeRepo)
	if err := u.ResolveDispute(context.Background(), "d-1", "admin-1", domain.OutcomeRefundBuyer, "x"); err == nil {
		t.Fatal("expected error because dispute is not UNDER_REVIEW")
	}
}

func TestDispute_GetDispute(t *testing.T) {
	disputeRepo := &mockDisputeRepo{
		onGetDisputeByID: func(ctx context.Context, id string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: "d-1", TransactionID: "tx-1", Status: domain.DisputeStatusOpen, Reason: "bad"}, nil
		},
	}
	u := usecase.NewDisputeUsecase(newMockUnitOfWorkWithDisputes(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, disputeRepo), disputeRepo)
	d, err := u.GetDispute(context.Background(), "d-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if d == nil || d.Reason != "bad" {
		t.Errorf("unexpected dispute: %+v", d)
	}
}
