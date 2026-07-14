package usecase_test

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// 🧠 1. MOCK REPOSITORY IMPLEMENTATIONS
// ============================================================================

type mockTransactionRepo struct {
	domain.TransactionRepository
	onCreateTransaction     func(ctx context.Context, tx *domain.Transaction) error
	onGetTransactionByID    func(ctx context.Context, id string) (*domain.Transaction, error)
	onUpdateTransactionStatus func(ctx context.Context, id string, status domain.TransactionStatus) error
}

func (m *mockTransactionRepo) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	if m.onCreateTransaction != nil {
		return m.onCreateTransaction(ctx, tx)
	}
	return nil
}

func (m *mockTransactionRepo) GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error) {
	if m.onGetTransactionByID != nil {
		return m.onGetTransactionByID(ctx, id)
	}
	return &domain.Transaction{}, nil
}

func (m *mockTransactionRepo) UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	if m.onUpdateTransactionStatus != nil {
		return m.onUpdateTransactionStatus(ctx, id, status)
	}
	return nil
}

type mockWalletRepo struct {
	domain.WalletRepository
	onUpdateBalanceTx func(ctx context.Context, txRecord *domain.RekberPayTransaction, modifier int64) error
	onExecuteInTx     func(ctx context.Context, fn func(txRepo domain.WalletRepository) error) error
}

func (m *mockWalletRepo) UpdateBalanceTx(ctx context.Context, txRecord *domain.RekberPayTransaction, modifier int64) error {
	if m.onUpdateBalanceTx != nil {
		return m.onUpdateBalanceTx(ctx, txRecord, modifier)
	}
	return nil
}

func (m *mockWalletRepo) ExecuteInTransaction(ctx context.Context, fn func(txRepo domain.WalletRepository) error) error {
	if m.onExecuteInTx != nil {
		return m.onExecuteInTx(ctx, fn)
	}
	return fn(m)
}

type mockFinanceRepo struct {
	domain.FinanceRepository
	onUpdatePlatformFinance func(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error
}

func (m *mockFinanceRepo) UpdatePlatformFinance(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error {
	if m.onUpdatePlatformFinance != nil {
		return m.onUpdatePlatformFinance(ctx, escrowDelta, revenueDelta, midtransFeeDelta)
	}
	return nil
}

// ============================================================================
// 🧪 2. UNIT TEST CASES (PARALLEL & MIRRORED STRUCTURE)
// ============================================================================

func TestLockFundsGoods_Success(t *testing.T) {
	ctx := context.Background()
	
	// Init mock repos
	txRepo := &mockTransactionRepo{}
	walletRepo := &mockWalletRepo{}
	financeRepo := &mockFinanceRepo{}
	calc := usecase.NewFinanceCalculator()
	
	u := usecase.NewTransactionGoodsUsecase(txRepo, walletRepo, financeRepo, calc)

	// Panggil target fungsi berdasarkan parameter referensi teranyar
	tx, err := u.LockFundsGoods(
		ctx,
		"buyer-uuid-123",
		"seller-uuid-456",
		100000,              // amountBase
		true,                // isRekberPay
		"BRONZE",            // sellerTier
		10000,               // shippingFee
		"REKBERPAY",         // paymentMethod
		"idem-key-goods-01", // idempotencyKey
	)

	if err != nil {
		t.Fatalf("Ekspektasi tidak ada error, namun dapet: %v", err)
	}

	if tx.Status != domain.StatusWaitingPayment {
		t.Errorf("Ekspektasi status %s, dapet: %s", domain.StatusWaitingPayment, tx.Status)
	}

	if tx.AmountGross <= tx.AmountBase {
		t.Errorf("Kalkulasi salah, amount gross (%d) harusnya bertambah biaya layanan", tx.AmountGross)
	}
}

func TestLockFundsGoods_InvalidAmount(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewTransactionGoodsUsecase(&mockTransactionRepo{}, &mockWalletRepo{}, &mockFinanceRepo{}, usecase.NewFinanceCalculator())

	_, err := u.LockFundsGoods(ctx, "b-id", "s-id", 0, true, "BRONZE", 10000, "REKBERPAY", "idem-02")
	if err == nil {
		t.Fatal("Ekspektasi error karena nominal transaksi barang = 0, tetapi dapet nil")
	}
}

func TestConfirmPaymentGoods_Success(t *testing.T) {
	ctx := context.Background()
	txID := "tx-uuid-goods-abc"

	txRepo := &mockTransactionRepo{
		onGetTransactionByID: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:          txID,
				BuyerID:     "buyer-uuid",
				Status:      domain.StatusWaitingPayment,
				AmountGross: 112500,
				ServiceFee:  2500,
			}, nil
		},
		onUpdateTransactionStatus: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			if status != domain.StatusFundsLocked {
				t.Errorf("Ekspektasi status transisi ke FUNDS_LOCKED, dapet: %s", status)
			}
			return nil
		},
	}

	walletRepo := &mockWalletRepo{}
	financeRepo := &mockFinanceRepo{}
	u := usecase.NewTransactionGoodsUsecase(txRepo, walletRepo, financeRepo, usecase.NewFinanceCalculator())

	err := u.ConfirmPaymentGoods(ctx, txID)
	if err != nil {
		t.Fatalf("Ekspektasi pembayaran sukses terkonfirmasi, dapet error: %v", err)
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
				t.Errorf("Ekspektasi status transisi ke RELEASED, dapet: %s", status)
			}
			return nil
		},
	}

	walletRepo := &mockWalletRepo{}
	financeRepo := &mockFinanceRepo{}
	u := usecase.NewTransactionGoodsUsecase(txRepo, walletRepo, financeRepo, usecase.NewFinanceCalculator())

	err := u.ReleaseFundsGoods(ctx, txID)
	if err != nil {
		t.Fatalf("Ekspektasi pelepasan dana sukses, dapet error: %v", err)
	}
}