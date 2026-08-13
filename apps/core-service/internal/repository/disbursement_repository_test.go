package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"rekberkuy/core-service/internal/domain"
)

func TestTransactionRepository_GetEventVendorPayoutByID(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "transaction_id", "vendor_user_id", "vendor_name", "amount_requested", "status"}).
		AddRow("pay-1", "tx-1", nil, "Vendor Luar", 800000, domain.VendorPayoutPendingDisbursement)
	mock.ExpectQuery(`FROM event_vendor_payouts WHERE id = \$1`).WithArgs("pay-1").WillReturnRows(rows)

	p, err := r.GetEventVendorPayoutByID(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if p.Status != domain.VendorPayoutPendingDisbursement {
		t.Errorf("status = %s, want PENDING_DISBURSEMENT", p.Status)
	}
	if p.VendorName != "Vendor Luar" {
		t.Errorf("vendor name = %s, want 'Vendor Luar'", p.VendorName)
	}
}

func TestTransactionRepository_MarkEventVendorPayoutDisbursed(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectExec(`UPDATE event_vendor_payouts`).
		WithArgs(domain.VendorPayoutDisbursed, "admin-1", "pay-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.MarkEventVendorPayoutDisbursed(context.Background(), "pay-1", "admin-1"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}
