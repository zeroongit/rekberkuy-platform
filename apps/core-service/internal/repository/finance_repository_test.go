package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFinanceRepository_GetPlatformFinance_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	r := &financeRepository{db: db}
	mock.ExpectQuery(`FROM platform_finances`).WillReturnError(sql.ErrNoRows)
	f, err := r.GetPlatformFinance(context.Background())
	if err != nil {
		t.Fatalf("expected success (default), got: %v", err)
	}
	if f.TotalEscrowBalance != 0 {
		t.Errorf("expected zero value when empty, got %d", f.TotalEscrowBalance)
	}
}

func TestFinanceRepository_GetPlatformFinance_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &financeRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "total_escrow_balance", "total_revenue", "total_midtrans_fees", "updated_at"}).
		AddRow("GLOBAL_FINANCE_ID", int64(500000), int64(12500), int64(3000), now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, total_escrow_balance, total_revenue, total_midtrans_fees, updated_at FROM platform_finances LIMIT 1`)).WillReturnRows(rows)
	f, err := r.GetPlatformFinance(context.Background())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if f.TotalRevenue != 12500 {
		t.Errorf("revenue = %d, want 12500", f.TotalRevenue)
	}
}

func TestFinanceRepository_UpdatePlatformFinance(t *testing.T) {
	db, mock := newMockDB(t)
	r := &financeRepository{db: db}
	mock.ExpectExec(`INSERT INTO platform_finances`).WithArgs(int64(-1000), int64(250), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdatePlatformFinance(context.Background(), -1000, 250, 0); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}
