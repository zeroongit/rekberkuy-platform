package domain

import (
	"context"
	"time"
)

type WalletTxType string

const (
	TxTopUp        WalletTxType = "TOPUP"
	TxPayment      WalletTxType = "PAYMENT"
	TxReceiveFunds WalletTxType = "RECEIVE_FUNDS"
	TxWithdraw     WalletTxType = "WITHDRAW"
	TxRefund       WalletTxType = "REFUND"
)

type WalletTxStatus string

const (
	WalletStatusPending WalletTxStatus = "PENDING"
	WalletStatusSuccess WalletTxStatus = "SUCCESS"
	WalletStatusFailed  WalletTxStatus = "FAILED"
)

type RekberPayWallet struct {
	UserID      string      `gorm:"type:uuid;primaryKey;not null" json:"user_id"`
	UserProfile UserProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Balance     int64       `gorm:"type:bigint;not null;default:0" json:"balance"`
	IsFrozen    bool        `gorm:"type:boolean;not null;default:false" json:"is_frozen"`
	UpdatedAt   time.Time   `gorm:"default:now()" json:"updated_at"`
}

type RekberPayTransaction struct {
	ID                     string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WalletID               string          `gorm:"type:uuid;not null;index" json:"wallet_id"`
	Wallet                 RekberPayWallet `gorm:"foreignKey:WalletID;constraint:OnDelete:CASCADE"`
	Type                   WalletTxType    `gorm:"type:varchar(50);not null" json:"type"`
	Status                 WalletTxStatus  `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	Amount                 int64           `gorm:"type:bigint;not null" json:"amount"`
	AdminFee               int64           `gorm:"type:bigint;not null;default:0" json:"admin_fee"`
	PlatformNetProfit      int64           `gorm:"type:bigint;not null;default:0" json:"platform_net_profit"`
	ReferenceTransactionID *string         `gorm:"type:uuid" json:"reference_transaction_id,omitempty"`
	MidtransTopUpID        *string         `gorm:"type:varchar(255)" json:"midtrans_topup_id,omitempty"`
	Description            *string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt              time.Time       `gorm:"default:now()" json:"created_at"`
}

type IdempotencyRecord struct {
	ID             string    `gorm:"type:varchar(255);primaryKey" json:"id"`
	RequestPath    string    `gorm:"type:varchar(255);not null" json:"request_path"`
	ResponseBody   []byte    `gorm:"type:bytea" json:"response_body,omitempty"`
	ResponseStatus int       `gorm:"type:integer;not null;default:0" json:"response_status"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
}

type IdempotencyRepository interface {
	CheckOrLock(ctx context.Context, record *IdempotencyRecord) (*IdempotencyRecord, bool, error)
	// SaveResponse stores the response (status + body) from the first request execution
	// so that duplicate requests (retry/double-click) with the same key get an identical response.
	SaveResponse(ctx context.Context, id string, status int, body []byte) error
}

type WalletRepository interface {
	GetBalance(ctx context.Context, userID string) (*RekberPayWallet, error)
	CreateWallet(ctx context.Context, userID string) error
	UpdateBalanceTx(ctx context.Context, txRecord *RekberPayTransaction, amountModifier int64) error
	GetAllUsersForCRMEvaluation(ctx context.Context) ([]*UserProfile, error)
	GetCRMLoyaltyByUserID(ctx context.Context, userID string) (*CRMLoyalty, error)
	UpdateCRMLoyalty(ctx context.Context, crmProfile *CRMLoyalty) error
	GetVendorAllocationsByTxID(ctx context.Context, transactionID string) ([]*EventVendorAllocation, error)
	// MarkVendorAllocationClaimed records that a vendor invoice drew down (part
	// of) the vendor's pledge: increments actual_paid_amount and flips the
	// status to CLAIMED once the pledge is fully consumed. Keeps pledge
	// reservations from blocking future invoices after they are settled.
	MarkVendorAllocationClaimed(ctx context.Context, transactionID, vendorID string, amount int64) error
	CreateVendorPayoutRecord(ctx context.Context, payout *EventVendorPayout) error
	GetVendorPayoutByID(ctx context.Context, payoutID string) (*EventVendorPayout, error)
	UpdateVendorPayoutStatus(ctx context.Context, payoutID string, status string) error
	// GetWalletTxByMidtransOrderID fetches the top-up draft transaction by its Midtrans order id.
	GetWalletTxByMidtransOrderID(ctx context.Context, orderID string) (*RekberPayTransaction, error)
	// MarkWalletTxStatusByOrderID updates the status of the top-up transaction row (PENDING -> SUCCESS/FAILED).
	MarkWalletTxStatusByOrderID(ctx context.Context, orderID string, status WalletTxStatus) error
	// GetWalletTxHistory returns the wallet ledger (newest first) for a user.
	GetWalletTxHistory(ctx context.Context, userID string, limit, offset int) ([]RekberPayTransaction, error)

	// Withdrawal lifecycle. The bank transfer itself happens out-of-band
	// (same stance as ADR-0002); these records track the request state.
	CreateWithdrawalRequest(ctx context.Context, w *WithdrawalRequest) error
	GetWithdrawalByID(ctx context.Context, id string) (*WithdrawalRequest, error)
	ListWithdrawalsByUser(ctx context.Context, userID string, limit, offset int) ([]WithdrawalRequest, error)
	ListPendingWithdrawals(ctx context.Context, limit, offset int) ([]WithdrawalRequest, error)
	// MarkWithdrawalDisbursed records that the bank transfer completed and the
	// admin confirmed it. actualCost (nullable) persists the REAL Midtrans fee
	// for the ledger true-up; nil keeps the booked estimate.
	MarkWithdrawalDisbursed(ctx context.Context, id string, adminID string, actualCost *int64) error
}

// WithdrawalStatus tracks a wallet withdrawal request. The wallet balance is
// debited immediately (amount + WithdrawFeeToUser); the bank payout completes
// out-of-band and is confirmed by an admin.
type WithdrawalStatus string

const (
	WithdrawalPending WithdrawalStatus = "PENDING"
	WithdrawalPaid    WithdrawalStatus = "PAID"
)

type WithdrawalRequest struct {
	ID             string           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         string           `gorm:"type:uuid;not null;index" json:"user_id"`
	UserProfile    UserProfile      `gorm:"foreignKey:UserID"`
	Amount         int64            `gorm:"type:bigint;not null" json:"amount"`
	Fee            int64            `gorm:"type:bigint;not null" json:"fee"` // GROSS user-facing fee (WithdrawFeeToUser)
	MidtransCost   int64            `gorm:"type:bigint;not null;default:0" json:"midtrans_cost"`
	MidtransCostActual *int64       `gorm:"type:bigint" json:"midtrans_cost_actual,omitempty"` // real cost, set at true-up; nil = still estimated
	BankName       string           `gorm:"type:varchar(100);not null" json:"bank_name"`
	AccountNumber  string           `gorm:"type:varchar(100);not null" json:"account_number"`
	AccountHolder  string           `gorm:"type:varchar(255);not null" json:"account_holder"`
	Status         WithdrawalStatus `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	ProcessedBy    *string          `gorm:"type:uuid" json:"processed_by,omitempty"`
	ProcessedAt    *time.Time       `json:"processed_at,omitempty"`
	CreatedAt      time.Time        `gorm:"default:now()" json:"created_at"`
	UpdatedAt      time.Time        `gorm:"default:now()" json:"updated_at"`
}
