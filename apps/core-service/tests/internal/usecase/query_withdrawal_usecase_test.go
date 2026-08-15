package usecase_test

import (
	"context"
	"errors"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// TRANSACTION QUERY / WITHDRAWAL USECASES — UNIT TESTS
// ============================================================================

func TestGetTransactionDetail_PartyCheck(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: id, BuyerID: "buyer-1", SellerID: "seller-1", Type: domain.TypeGoods}, nil
		},
		onGetGoodsDetail: func(ctx context.Context, txID string) (*domain.TransactionGoods, error) {
			return &domain.TransactionGoods{TransactionID: txID, ShippingCourier: "JNE"}, nil
		},
	}
	u := usecase.NewTransactionQueryUsecase(txRepo)

	detail, err := u.GetTransactionDetail(ctx, "buyer-1", "tx-1")
	if err != nil {
		t.Fatalf("buyer must see the detail, got: %v", err)
	}
	if detail.Goods == nil || detail.Goods.ShippingCourier != "JNE" {
		t.Error("goods detail must be composed into the view")
	}

	if _, err := u.GetTransactionDetail(ctx, "stranger-1", "tx-1"); err != usecase.ErrNotTransactionParty {
		t.Fatalf("third party must be refused, got: %v", err)
	}
}

func TestGetTransactionDetail_ServicesComposesMilestones(t *testing.T) {
	ctx := context.Background()
	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{ID: id, BuyerID: "b", SellerID: "s", Type: domain.TypeServices}, nil
		},
		onGetServicesDetail: func(ctx context.Context, txID string) (*domain.TransactionServices, error) {
			return &domain.TransactionServices{TransactionID: txID}, nil
		},
		onGetMilestones: func(ctx context.Context, txID string) ([]domain.ServiceMilestone, error) {
			return []domain.ServiceMilestone{{ID: "ms-1"}, {ID: "ms-2"}}, nil
		},
	}
	u := usecase.NewTransactionQueryUsecase(txRepo)

	detail, err := u.GetTransactionDetail(ctx, "s", "tx-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if detail.Services == nil {
		t.Error("services detail must be composed")
	}
	if len(detail.Milestones) != 2 {
		t.Errorf("milestones = %d, want 2", len(detail.Milestones))
	}
}

func TestRequestWithdrawal_Success(t *testing.T) {
	ctx := context.Background()

	var debited *domain.RekberPayTransaction
	var debitMod int64
	var created *domain.WithdrawalRequest
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error {
			debited, debitMod = rec, mod
			return nil
		},
		onCreateWithdrawal: func(ctx context.Context, w *domain.WithdrawalRequest) error {
			created = w
			return nil
		},
	}
	var finEscrow, finRevenue, finMidtrans int64
	finRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error {
			finEscrow, finRevenue, finMidtrans = escrowDelta, revenueDelta, midtransFeeDelta
			return nil
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, finRepo, nil)
	u := usecase.NewWithdrawalUsecase(uow, walletRepo, &mockDisbursementClient{estimate: domain.WithdrawFeeMidtransCost})

	req, err := u.RequestWithdrawal(ctx, "user-1", 500000, "BCA", "1234567890", "Budi Santoso")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if debited == nil {
		t.Fatal("wallet must be debited")
	}
	if debited.Type != domain.TxWithdraw {
		t.Errorf("wallet tx type = %s, want WITHDRAW", debited.Type)
	}
	if debited.AdminFee != domain.WithdrawFeeToUser {
		t.Errorf("admin fee = %d, want %d", debited.AdminFee, domain.WithdrawFeeToUser)
	}
	if debitMod != -500000 {
		t.Errorf("debit modifier = %d, want -500000 (fee deducted separately by UpdateBalanceTx)", debitMod)
	}
	if created == nil || created.Status != domain.WithdrawalPending {
		t.Error("a PENDING withdrawal request must be created")
	}
	if req.Amount != 500000 || req.Fee != domain.WithdrawFeeToUser {
		t.Errorf("request amount/fee = %d/%d, want 500000/%d", req.Amount, req.Fee, domain.WithdrawFeeToUser)
	}
	// Fee split: the gross fee covers the Midtrans disbursement cost; only the
	// remainder is net platform profit.
	if finMidtrans != domain.WithdrawFeeMidtransCost {
		t.Errorf("midtrans fee booked = %d, want %d", finMidtrans, domain.WithdrawFeeMidtransCost)
	}
	wantRevenue := int64(domain.WithdrawFeeToUser - domain.WithdrawFeeMidtransCost)
	if finRevenue != wantRevenue {
		t.Errorf("revenue booked = %d, want %d (gross fee minus midtrans cost)", finRevenue, wantRevenue)
	}
	if finEscrow != 0 {
		t.Errorf("escrow delta = %d, want 0 (wallet balance is not escrow)", finEscrow)
	}
}

func TestRequestWithdrawal_InsufficientBalance(t *testing.T) {
	ctx := context.Background()
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error {
			return context.DeadlineExceeded // simulate repo refusing (e.g. insufficient balance)
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, &mockFinanceRepo{}, nil)
	u := usecase.NewWithdrawalUsecase(uow, walletRepo, nil)

	if _, err := u.RequestWithdrawal(ctx, "user-1", 500000, "BCA", "123", "Budi"); err == nil {
		t.Fatal("expected error when the wallet debit is refused")
	}
}

func TestRequestWithdrawal_Validation(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewWithdrawalUsecase(newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, nil), &mockWalletRepo{}, nil)

	if _, err := u.RequestWithdrawal(ctx, "u", 0, "BCA", "123", "Budi"); !errors.Is(err, usecase.ErrInvalidWithdrawalReqt) {
		t.Errorf("zero amount must be rejected as ErrInvalidWithdrawalReqt, got: %v", err)
	}
	if _, err := u.RequestWithdrawal(ctx, "u", 1000, "", "123", "Budi"); !errors.Is(err, usecase.ErrInvalidWithdrawalReqt) {
		t.Errorf("missing bank name must be rejected as ErrInvalidWithdrawalReqt, got: %v", err)
	}
}

func TestRequestWithdrawal_DatabaseErrorIsNotValidation(t *testing.T) {
	ctx := context.Background()
	walletRepo := &mockWalletRepo{
		onUpdateBalanceTx: func(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error {
			return errors.New("connection refused")
		},
	}
	u := usecase.NewWithdrawalUsecase(newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, &mockFinanceRepo{}, nil), walletRepo, nil)

	_, err := u.RequestWithdrawal(ctx, "u", 1000, "BCA", "123", "Budi")
	if err == nil {
		t.Fatal("expected error on DB failure")
	}
	if errors.Is(err, usecase.ErrInvalidWithdrawalReqt) {
		t.Error("a DB failure must NOT be classified as a validation error (handler maps it to 500)")
	}
}

func TestMarkWithdrawalDisbursed(t *testing.T) {
	ctx := context.Background()

	marked := false
	walletRepo := &mockWalletRepo{
		onGetWithdrawalByID: func(ctx context.Context, id string) (*domain.WithdrawalRequest, error) {
			return &domain.WithdrawalRequest{ID: id, Status: domain.WithdrawalPending}, nil
		},
		onMarkWithdrawalDisbursed: func(ctx context.Context, id string, adminID string, actualCost *int64) error {
			marked = true
			return nil
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, &mockFinanceRepo{}, nil)
	u := usecase.NewWithdrawalUsecase(uow, walletRepo, nil)

	if err := u.MarkWithdrawalDisbursed(ctx, "wd-1", "admin-1", nil); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !marked {
		t.Error("MarkWithdrawalDisbursed was not called")
	}
}

func TestMarkWithdrawalDisbursed_AlreadyPaid(t *testing.T) {
	ctx := context.Background()
	walletRepo := &mockWalletRepo{
		onGetWithdrawalByID: func(ctx context.Context, id string) (*domain.WithdrawalRequest, error) {
			return &domain.WithdrawalRequest{ID: id, Status: domain.WithdrawalPaid}, nil
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, &mockFinanceRepo{}, nil)
	u := usecase.NewWithdrawalUsecase(uow, walletRepo, nil)

	if err := u.MarkWithdrawalDisbursed(ctx, "wd-1", "admin-1", nil); err != usecase.ErrWithdrawalAlreadyPaid {
		t.Fatalf("expected ErrWithdrawalAlreadyPaid, got: %v", err)
	}
}

// --- True-up: real Midtrans cost reconciliation --------------------------------

func TestMarkWithdrawalDisbursed_TrueUpBooksDelta(t *testing.T) {
	ctx := context.Background()

	// Booked estimate 4.000 at request time; the real cost turned out 5.500.
	// The true-up must shift the 1.500 delta from revenue to the Midtrans ledger.
	var finRevenue, finMidtrans int64
	walletRepo := &mockWalletRepo{
		onGetWithdrawalByID: func(ctx context.Context, id string) (*domain.WithdrawalRequest, error) {
			return &domain.WithdrawalRequest{ID: id, Status: domain.WithdrawalPending, MidtransCost: 4000}, nil
		},
		onMarkWithdrawalDisbursed: func(ctx context.Context, id string, adminID string, actualCost *int64) error {
			if actualCost == nil || *actualCost != 5500 {
				t.Errorf("actual cost passed to repo = %v, want 5500", actualCost)
			}
			return nil
		},
	}
	finRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error {
			finRevenue, finMidtrans = revenueDelta, midtransFeeDelta
			return nil
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, finRepo, nil)
	u := usecase.NewWithdrawalUsecase(uow, walletRepo, nil)

	actual := int64(5500)
	if err := u.MarkWithdrawalDisbursed(ctx, "wd-1", "admin-1", &actual); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if finMidtrans != 1500 {
		t.Errorf("midtrans delta = %d, want +1500", finMidtrans)
	}
	if finRevenue != -1500 {
		t.Errorf("revenue delta = %d, want -1500 (estimate replaced by real cost)", finRevenue)
	}
}

func TestMarkWithdrawalDisbursed_TrueUpNoDeltaWhenEqual(t *testing.T) {
	ctx := context.Background()

	financeTouched := false
	walletRepo := &mockWalletRepo{
		onGetWithdrawalByID: func(ctx context.Context, id string) (*domain.WithdrawalRequest, error) {
			return &domain.WithdrawalRequest{ID: id, Status: domain.WithdrawalPending, MidtransCost: 4000}, nil
		},
	}
	finRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, e, r, m2 int64) error {
			financeTouched = true
			return nil
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, finRepo, nil)
	u := usecase.NewWithdrawalUsecase(uow, walletRepo, nil)

	actual := int64(4000)
	if err := u.MarkWithdrawalDisbursed(ctx, "wd-1", "admin-1", &actual); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if financeTouched {
		t.Error("no finance entry should be written when the real cost equals the booked estimate")
	}
}

func TestRequestWithdrawal_UsesProviderEstimate(t *testing.T) {
	ctx := context.Background()

	// When the (future) provider adapter reports a live estimate, the booking
	// uses it instead of the static default.
	var bookedRevenue, bookedMidtrans int64
	walletRepo := &mockWalletRepo{}
	finRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, e, r, m2 int64) error {
			bookedRevenue, bookedMidtrans = r, m2
			return nil
		},
	}
	client := &mockDisbursementClient{
		onEstimate: func(ctx context.Context, req domain.DisbursementRequest) (int64, error) {
			return 5000, nil
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, walletRepo, finRepo, nil)
	u := usecase.NewWithdrawalUsecase(uow, walletRepo, client)

	req, err := u.RequestWithdrawal(ctx, "user-1", 500000, "BCA", "123", "Budi")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if req.MidtransCost != 5000 {
		t.Errorf("booked estimate = %d, want 5000 from provider", req.MidtransCost)
	}
	if bookedMidtrans != 5000 || bookedRevenue != domain.WithdrawFeeToUser-5000 {
		t.Errorf("booking = revenue %d / midtrans %d, want %d/5000", bookedRevenue, bookedMidtrans, domain.WithdrawFeeToUser-5000)
	}
}

func TestRequestWithdrawal_EstimateFailureFallsBack(t *testing.T) {
	ctx := context.Background()

	// An estimation failure must never block a withdrawal: fall back to the
	// configured default and continue.
	var bookedMidtrans int64
	finRepo := &mockFinanceRepo{
		onUpdatePlatformFinance: func(ctx context.Context, e, r, m2 int64) error {
			bookedMidtrans = m2
			return nil
		},
	}
	client := &mockDisbursementClient{
		onEstimate: func(ctx context.Context, req domain.DisbursementRequest) (int64, error) {
			return 0, errors.New("provider unreachable")
		},
	}
	uow := newMockUnitOfWork(&mockTransactionRepo{}, &mockWalletRepo{}, finRepo, nil)
	u := usecase.NewWithdrawalUsecase(uow, &mockWalletRepo{}, client)

	if _, err := u.RequestWithdrawal(ctx, "user-1", 500000, "BCA", "123", "Budi"); err != nil {
		t.Fatalf("withdrawal must proceed despite estimation failure, got: %v", err)
	}
	if bookedMidtrans != domain.WithdrawFeeMidtransCost {
		t.Errorf("fallback booking = %d, want default %d", bookedMidtrans, domain.WithdrawFeeMidtransCost)
	}
}
