package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"rekberkuy/core-service/internal/domain"
)

func TestTransactionRepository_CreateTransaction(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectExec(`INSERT INTO transactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	tx := &domain.Transaction{ID: "t1", BuyerID: "b", SellerID: "s", Type: domain.TypeGoods, Status: domain.StatusWaitingPayment}
	if err := r.CreateTransaction(context.Background(), tx); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestTransactionRepository_GetTransactionByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "buyer_id", "seller_id", "type", "status", "amount_base", "shipping_fee",
		"service_fee", "midtrans_fee", "amount_gross", "amount_net", "midtrans_order_id", "idempotency_key",
		"payment_method", "blockchain_tx_hash", "blockchain_logged_at", "created_at", "updated_at"}).
		AddRow("t1", "b", "s", "GOODS", "FUNDS_LOCKED", 100, 0, 0, 0, 100, 90, "ord", "idem", "REKBERPAY", nil, nil, now, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, buyer_id, seller_id, type, status, amount_base, shipping_fee,`)).WithArgs("t1").WillReturnRows(rows)
	tx, err := r.GetTransactionByID(context.Background(), "t1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if tx.Status != domain.StatusFundsLocked {
		t.Errorf("status = %s, want FUNDS_LOCKED", tx.Status)
	}
}

func TestTransactionRepository_GetTransactionByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectQuery(`FROM transactions`).WithArgs("ghost").WillReturnError(sql.ErrNoRows)
	if _, err := r.GetTransactionByID(context.Background(), "ghost"); err == nil {
		t.Fatal("expected error because transaction not found")
	}
}

func TestTransactionRepository_UpdateTransactionStatus(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectExec(`UPDATE transactions SET status = \$1, updated_at = NOW`).
		WithArgs(domain.StatusReleased, "t1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateTransactionStatus(context.Background(), "t1", domain.StatusReleased); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestTransactionRepository_UpdateBlockchainLog(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectExec(`UPDATE transactions SET blockchain_tx_hash = \$1`).
		WithArgs("0xhash", "t1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateBlockchainLog(context.Background(), "t1", "0xhash"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestTransactionRepository_GetTransactionByMidtransOrderID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectQuery(`WHERE midtrans_order_id = \$1`).WithArgs("ghost-order").WillReturnError(sql.ErrNoRows)
	if _, err := r.GetTransactionByMidtransOrderID(context.Background(), "ghost-order"); err == nil {
		t.Fatal("expected error because order id not found")
	}
}

func TestTransactionRepository_GetExpiredLockedTransactions(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	rows := sqlmock.NewRows([]string{"id"}).AddRow("t-a").AddRow("t-b")
	mock.ExpectQuery(`WHERE t.status = 'FUNDS_LOCKED'`).WillReturnRows(rows)
	ids, err := r.GetExpiredLockedTransactions(context.Background())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("len = %d, want 2", len(ids))
	}
}

func TestTransactionRepository_UpdateMilestoneStatus(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectExec(`UPDATE service_milestones SET status`).WithArgs("RELEASED", "m1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateMilestoneStatus(context.Background(), "m1", "RELEASED"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestTransactionRepository_UpdateEventVendorPayoutStatus(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectExec(`UPDATE event_vendor_payouts SET status`).WithArgs(domain.VendorPayoutApproved, "p1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateEventVendorPayoutStatus(context.Background(), "p1", domain.VendorPayoutApproved); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}
