package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"rekberkuy/core-service/internal/domain"
)

var now = time.Now()

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestWalletRepository_GetBalance_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}

	rows := sqlmock.NewRows([]string{"user_id", "balance", "is_frozen", "updated_at"}).
		AddRow("u-1", int64(50000), false, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, balance, is_frozen, updated_at FROM rekberpay_wallets WHERE user_id = $1`)).
		WithArgs("u-1").WillReturnRows(rows)

	w, err := r.GetBalance(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if w.Balance != 50000 {
		t.Errorf("balance = %d, want 50000", w.Balance)
	}
}

func TestWalletRepository_GetBalance_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, balance, is_frozen, updated_at FROM rekberpay_wallets WHERE user_id = $1`)).
		WithArgs("ghost").WillReturnError(sql.ErrNoRows)

	if _, err := r.GetBalance(context.Background(), "ghost"); err == nil {
		t.Fatal("expected error because wallet not found")
	}
}

func TestWalletRepository_CreateWallet(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO rekberpay_wallets (user_id, balance, is_frozen, updated_at) VALUES ($1, 0, false, NOW()) ON CONFLICT (user_id) DO NOTHING`)).
		WithArgs("u-new").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.CreateWallet(context.Background(), "u-new"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

// TestWalletRepository_UpdateBalanceTx_InsufficientBalance verifies the rejection
// when the balance is insufficient (standalone path: BeginTx + Rollback).
func TestWalletRepository_UpdateBalanceTx_InsufficientBalance(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance, is_frozen FROM rekberpay_wallets WHERE user_id = $1 FOR UPDATE`)).
		WithArgs("u-poor").WillReturnRows(sqlmock.NewRows([]string{"balance", "is_frozen"}).AddRow(int64(1000), false))
	mock.ExpectRollback()

	ref := "tx-1"
	rec := &domain.RekberPayTransaction{
		ID: "wt-1", WalletID: "u-poor", Type: domain.TxPayment,
		Status: domain.WalletStatusSuccess, Amount: 10000, AdminFee: 0,
		ReferenceTransactionID: &ref,
	}
	err := r.UpdateBalanceTx(context.Background(), rec, -10000)
	if err == nil {
		t.Fatal("expected error because balance is insufficient")
	}
}

// TestWalletRepository_UpdateBalanceTx_FrozenWallet rejects a mutation on a frozen wallet.
func TestWalletRepository_UpdateBalanceTx_FrozenWallet(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance, is_frozen FROM rekberpay_wallets WHERE user_id = $1 FOR UPDATE`)).
		WithArgs("u-frozen").WillReturnRows(sqlmock.NewRows([]string{"balance", "is_frozen"}).AddRow(int64(100000), true))
	mock.ExpectRollback()

	rec := &domain.RekberPayTransaction{ID: "wt", WalletID: "u-frozen", Type: domain.TxPayment, Status: domain.WalletStatusSuccess, Amount: 100}
	if err := r.UpdateBalanceTx(context.Background(), rec, -100); err == nil {
		t.Fatal("expected error because wallet is frozen")
	}
}

// TestWalletRepository_UpdateBalanceTx_Success credit (positive modifier).
func TestWalletRepository_UpdateBalanceTx_CreditSuccess(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance, is_frozen FROM rekberpay_wallets WHERE user_id = $1 FOR UPDATE`)).
		WithArgs("u-seller").WillReturnRows(sqlmock.NewRows([]string{"balance", "is_frozen"}).AddRow(int64(0), false))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE rekberpay_wallets SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2`)).
		WithArgs(int64(90000), "u-seller").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO rekberpay_transactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	desc := "kredit"
	rec := &domain.RekberPayTransaction{ID: "wt", WalletID: "u-seller", Type: domain.TxReceiveFunds, Status: domain.WalletStatusSuccess, Amount: 90000, Description: &desc}
	if err := r.UpdateBalanceTx(context.Background(), rec, 90000); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestWalletRepository_MarkWalletTxStatusByOrderID(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE rekberpay_transactions SET status = $1 WHERE midtrans_topup_id = $2`)).
		WithArgs(domain.WalletStatusFailed, "order-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.MarkWalletTxStatusByOrderID(context.Background(), "order-1", domain.WalletStatusFailed); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestWalletRepository_GetWalletTxByMidtransOrderID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	mock.ExpectQuery(`SELECT id, wallet_id, type, status, amount, admin_fee, platform_net_profit,`).WithArgs("ghost-order").
		WillReturnError(sql.ErrNoRows)
	_, err := r.GetWalletTxByMidtransOrderID(context.Background(), "ghost-order")
	if err == nil {
		t.Fatal("expected error because draft not found")
	}
}

func TestWalletRepository_UpdateBalanceTx_NestedTxNoCommit(t *testing.T) {
	// Path r.tx != nil (nested): does not perform Begin/Commit itself, uses the injected tx.
	db, mock := newMockDB(t)
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx failed: %v", err)
	}
	r := &walletRepository{db: db, tx: tx}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance, is_frozen FROM rekberpay_wallets WHERE user_id = $1 FOR UPDATE`)).
		WithArgs("u-seller").WillReturnRows(sqlmock.NewRows([]string{"balance", "is_frozen"}).AddRow(int64(0), false))
	mock.ExpectExec(`UPDATE rekberpay_wallets SET balance`).WithArgs(int64(90000), "u-seller").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO rekberpay_transactions`).WillReturnResult(sqlmock.NewResult(0, 1))

	desc := "kredit nested"
	rec := &domain.RekberPayTransaction{ID: "wt", WalletID: "u-seller", Type: domain.TxReceiveFunds, Status: domain.WalletStatusSuccess, Amount: 90000, Description: &desc}
	if err := r.UpdateBalanceTx(context.Background(), rec, 90000); err != nil {
		t.Fatalf("expected success (nested), got: %v", err)
	}
}

func TestUnitOfWork_CommitsOnSuccess(t *testing.T) {
	db, mock := newMockDB(t)
	uow := NewUnitOfWork(db)
	called := false
	mock.ExpectBegin()
	mock.ExpectCommit()
	if err := uow.Do(context.Background(), func(ctx context.Context, stores domain.TxStores) error {
		called = true
		if stores.Wallets == nil || stores.Transactions == nil || stores.Finance == nil || stores.Users == nil {
			t.Error("all stores must be populated with tx-scoped instances")
		}
		return nil
	}); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !called {
		t.Error("fn must be called")
	}
}

func TestUnitOfWork_RollsBackOnError(t *testing.T) {
	db, mock := newMockDB(t)
	uow := NewUnitOfWork(db)
	mock.ExpectBegin()
	mock.ExpectRollback()
	err := uow.Do(context.Background(), func(ctx context.Context, stores domain.TxStores) error {
		return errors.New("boom")
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected error 'boom', got: %v", err)
	}
}
