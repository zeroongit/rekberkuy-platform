package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"rekberkuy/core-service/internal/domain"
)

func TestKYCRepository_SubmitKYC(t *testing.T) {
	db, mock := newMockDB(t)
	r := &kycRepository{db: db}
	mock.ExpectExec(`INSERT INTO kyc_submissions`).
		WithArgs("kyc-1", "u-1", string(domain.RoleVerifiedMerchant), "3171", "https://ktp", "https://selfie", string(domain.KYCPending)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	kyc := &domain.KYCSubmission{
		ID: "kyc-1", UserID: "u-1", TargetRole: domain.RoleVerifiedMerchant,
		IDCardNumber: "3171", IDCardURL: "https://ktp", SelfieURL: "https://selfie", Status: domain.KYCPending,
	}
	if err := r.SubmitKYC(context.Background(), kyc); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestVendorRepository_CreateVendor(t *testing.T) {
	db, mock := newMockDB(t)
	r := &vendorRepository{db: db}
	mock.ExpectExec(`INSERT INTO vendor_profiles`).
		WithArgs("v-1", "Katering Sedap", "CATERING", false).
		WillReturnResult(sqlmock.NewResult(0, 1))
	v := &domain.VendorProfile{VendorID: "v-1", BusinessName: "Katering Sedap", Category: "CATERING", IsVerified: false}
	if err := r.CreateVendor(context.Background(), v); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

// --- Remaining wallet methods ---

func TestWalletRepository_GetCRMLoyaltyByUserID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	mock.ExpectQuery(`FROM crm_loyalty WHERE user_id`).WillReturnError(sqlNoRows())
	if _, err := r.GetCRMLoyaltyByUserID(context.Background(), "ghost"); err == nil {
		t.Fatal("expected error because CRM not found")
	}
}

func TestWalletRepository_UpdateCRMLoyalty(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	mock.ExpectExec(`UPDATE crm_loyalty SET current_tier`).WithArgs("GOLD", 0, "u-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	crm := &domain.CRMLoyalty{UserID: "u-1", CurrentTier: "GOLD", ConsecutiveFailedMonths: 0}
	if err := r.UpdateCRMLoyalty(context.Background(), crm); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestWalletRepository_GetVendorPayoutByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "transaction_id", "vendor_user_id", "vendor_name", "vendor_bank_name",
		"vendor_account_number", "amount_requested", "expense_description", "invoice_file_url", "payout_phase",
		"status", "is_disbursed_by_midtrans", "created_at"}).
		AddRow("p1", "tx1", nil, "Foto C", "BCA", "123", int64(5000), "jasa", "url", "DP", "PENDING", false, now)
	mock.ExpectQuery(`FROM event_vendor_payouts WHERE id = \$1`).WithArgs("p1").WillReturnRows(rows)
	p, err := r.GetVendorPayoutByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if p.VendorName != "Foto C" {
		t.Errorf("vendor = %s, want Foto C", p.VendorName)
	}
}

func TestWalletRepository_UpdateVendorPayoutStatus(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	mock.ExpectExec(`UPDATE event_vendor_payouts SET status`).WithArgs("APPROVED", "p1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.UpdateVendorPayoutStatus(context.Background(), "p1", "APPROVED"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestWalletRepository_CreateVendorPayoutRecord(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	mock.ExpectExec(`INSERT INTO event_vendor_payouts`).WillReturnResult(sqlmock.NewResult(0, 1))
	p := &domain.EventVendorPayout{ID: "p1", TransactionID: "tx1", VendorName: "V", VendorBankName: "B", VendorAccountNumber: "1", AmountRequested: 100, ExpenseDescription: "d", InvoiceFileURL: "u", PayoutPhase: "DP", Status: "PENDING"}
	if err := r.CreateVendorPayoutRecord(context.Background(), p); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestWalletRepository_GetVendorAllocationsByTxID(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "transaction_id", "vendor_id", "allocated_amount", "actual_paid_amount", "status", "created_at"}).
		AddRow("a1", "tx1", "v1", int64(1000), int64(0), "PLEDGED", now)
	mock.ExpectQuery(regexp.QuoteMeta(`FROM event_vendor_allocations`)).WithArgs("tx1").WillReturnRows(rows)
	allocs, err := r.GetVendorAllocationsByTxID(context.Background(), "tx1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(allocs) != 1 {
		t.Errorf("len = %d, want 1", len(allocs))
	}
}

func TestWalletRepository_GetAllUsersForCRMEvaluation(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "username", "full_name", "role", "phone_number", "created_at", "updated_at"}).
		AddRow("u1", "a", "A", "USER", nil, now, now)
	mock.ExpectQuery(`FROM user_profiles`).WillReturnRows(rows)
	users, err := r.GetAllUsersForCRMEvaluation(context.Background())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("len = %d, want 1", len(users))
	}
}

// --- Remaining transaction methods ---

func TestTransactionRepository_GetMilestoneByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectQuery(`FROM service_milestones WHERE id`).WillReturnError(sqlNoRows())
	if _, err := r.GetMilestoneByID(context.Background(), "ghost"); err == nil {
		t.Fatal("expected error because milestone not found")
	}
}

func TestTransactionRepository_GetEventVendorPayoutsByTxID_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	mock.ExpectQuery(`FROM event_vendor_payouts WHERE transaction_id`).WithArgs("tx1").WillReturnRows(sqlmock.NewRows([]string{"id", "transaction_id", "vendor_user_id", "vendor_name", "amount_requested", "status"}))
	out, err := r.GetEventVendorPayoutsByTxID(context.Background(), "tx1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("len = %d, want 0", len(out))
	}
}

func TestTransactionRepository_GetMilestoneByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "transaction_id", "milestone_index", "title", "amount", "status"}).
		AddRow("m1", "tx1", 1, "Desain", int64(5000), "PENDING")
	mock.ExpectQuery(`FROM service_milestones WHERE id`).WithArgs("m1").WillReturnRows(rows)
	m, err := r.GetMilestoneByID(context.Background(), "m1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if m.Title != "Desain" {
		t.Errorf("title = %s, want Desain", m.Title)
	}
}

func TestTransactionRepository_GetTransactionByMidtransOrderID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &TransactionRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "buyer_id", "seller_id", "type", "status", "amount_base", "shipping_fee",
		"service_fee", "midtrans_fee", "amount_gross", "amount_net", "midtrans_order_id", "idempotency_key",
		"payment_method", "blockchain_tx_hash", "blockchain_logged_at", "created_at", "updated_at"}).
		AddRow("t1", "b", "s", "GOODS", "WAITING_PAYMENT", 100, 0, 0, 0, 100, 90, "ord-1", "idem", "REKBERPAY", nil, nil, now, now)
	mock.ExpectQuery(`WHERE midtrans_order_id = \$1`).WithArgs("ord-1").WillReturnRows(rows)
	tx, err := r.GetTransactionByMidtransOrderID(context.Background(), "ord-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if tx.MidtransOrderID != "ord-1" {
		t.Errorf("order id = %s, want ord-1", tx.MidtransOrderID)
	}
}

func TestWalletRepository_GetWalletTxByMidtransOrderID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	r := &walletRepository{db: db}
	orderID := "ord-topup"
	rows := sqlmock.NewRows([]string{"id", "wallet_id", "type", "status", "amount", "admin_fee", "platform_net_profit",
		"reference_transaction_id", "midtrans_topup_id", "description", "created_at"}).
		AddRow("w1", "u-1", "TOPUP", "PENDING", int64(5000), 0, 0, nil, orderID, nil, now)
	mock.ExpectQuery(`WHERE midtrans_topup_id = \$1`).WithArgs(orderID).WillReturnRows(rows)
	w, err := r.GetWalletTxByMidtransOrderID(context.Background(), orderID)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if w.Amount != 5000 {
		t.Errorf("amount = %d, want 5000", w.Amount)
	}
}

// Nested-tx (r.tx != nil) path coverage for simple read/create methods.
func TestWalletRepository_NestedTx_ReadAndCreate(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	r := &walletRepository{db: db, tx: tx}

	mock.ExpectQuery(`SELECT user_id, balance, is_frozen, updated_at FROM rekberpay_wallets`).
		WithArgs("u-1").WillReturnRows(sqlmock.NewRows([]string{"user_id", "balance", "is_frozen", "updated_at"}).AddRow("u-1", int64(100), false, now))
	if _, err := r.GetBalance(context.Background(), "u-1"); err != nil {
		t.Fatalf("GetBalance nested failed: %v", err)
	}

	mock.ExpectExec(`INSERT INTO rekberpay_wallets`).WithArgs("u-new").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.CreateWallet(context.Background(), "u-new"); err != nil {
		t.Fatalf("CreateWallet nested failed: %v", err)
	}
}

// helper: sql.ErrNoRows without re-importing it in every file
func sqlNoRows() error { return errNoRows }

var errNoRows = errNoRowsVal{}

type errNoRowsVal struct{}

func (errNoRowsVal) Error() string { return "sql: no rows in result set" }
