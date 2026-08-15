package domain

import (
	"context"
	"time"
)

const (
	MaxMemberEventLimit  = 10000000
	QRISTotalFeePercent  = 0.01
	WithdrawFeeToUser    = 7500
	FeeGoodsRekberPay    = 2500
	FeeGoodsNonRekberPay = 5000
)

type RekberType string
type TransactionStatus string

const (
	TypeGoods    RekberType = "GOODS"
	TypeServices RekberType = "SERVICES"
	TypeEvents   RekberType = "EVENTS"

	StatusWaitingPayment TransactionStatus = "WAITING_PAYMENT"
	StatusFundsLocked    TransactionStatus = "FUNDS_LOCKED"
	StatusDisputed       TransactionStatus = "DISPUTED"
	StatusReleased       TransactionStatus = "RELEASED"
	StatusRefunded       TransactionStatus = "REFUNDED"
)

// Status for EventVendorPayout (event vendor invoice).
const (
	VendorPayoutPending             = "PENDING"              // Awaiting review/disbursement
	VendorPayoutApproved            = "APPROVED"             // Disbursed (internal wallet credit)
	VendorPayoutPendingDisbursement = "PENDING_DISBURSEMENT" // External vendor, awaiting bank disbursement
	VendorPayoutDisbursed           = "DISBURSED"            // External vendor paid out-of-band; terminal
)

// Status for EventVendorAllocation (the EO's vendor pledge).
const (
	VendorAllocationPledged  = "PLEDGED"  // Reserved against the escrow, invoice not yet claimed
	VendorAllocationClaimed  = "CLAIMED"  // A vendor invoice has drawn down (part of) the pledge
)

type Transaction struct {
	ID                 string            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BuyerID            string            `gorm:"type:uuid;not null;index" json:"buyer_id"`
	SellerID           string            `gorm:"type:uuid;not null;index" json:"seller_id"`
	Buyer              UserProfile       `gorm:"foreignKey:BuyerID"`
	Seller             UserProfile       `gorm:"foreignKey:SellerID"`
	Type               RekberType        `gorm:"type:varchar(50);not null" json:"type"`
	Status             TransactionStatus `gorm:"type:varchar(50);not null;default:'WAITING_PAYMENT'" json:"status"`
	AmountBase         int64             `gorm:"type:bigint;not null" json:"amount_base"`
	ShippingFee        int64             `gorm:"type:bigint;not null;default:0" json:"shipping_fee"`
	ServiceFee         int64             `gorm:"type:bigint;not null" json:"service_fee"`
	MidtransFee        int64             `gorm:"type:bigint;not null" json:"midtrans_fee"`
	AmountGross        int64             `gorm:"type:bigint;not null" json:"amount_gross"`
	AmountNet          int64             `gorm:"type:bigint;not null" json:"amount_net"`
	MidtransOrderID    string            `gorm:"type:varchar(255);not null;unique" json:"midtrans_order_id"`
	IdempotencyKey     string            `gorm:"type:varchar(255);not null;unique" json:"idempotency_key"`
	PaymentMethod      string            `gorm:"type:varchar(100);not null" json:"payment_method"`
	BlockchainTxHash   *string           `gorm:"type:varchar(255);unique" json:"blockchain_tx_hash,omitempty"`
	BlockchainLoggedAt *time.Time        `json:"blockchain_logged_at,omitempty"`
	CreatedAt          time.Time         `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time         `gorm:"default:now()" json:"updated_at"`
}

type TransactionGoods struct {
	TransactionID          string              `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	Transaction            Transaction         `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	SubSubCategoryID       uint64              `gorm:"not null" json:"sub_sub_category_id"`
	SubSubCategory         GoodsSubSubCategory `gorm:"foreignKey:SubSubCategoryID" json:"-"`
	ShippingCourier        string              `gorm:"type:varchar(100);not null" json:"shipping_courier"`
	ShippingTrackingNumber *string             `gorm:"type:varchar(255)" json:"shipping_tracking_number,omitempty"`
	ShippingAddress        string              `gorm:"type:text;not null" json:"shipping_address"`
	AutoConfirmDeadline    time.Time           `gorm:"not null" json:"auto_confirm_deadline"`
}

type TransactionServices struct {
	TransactionID    string                `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	Transaction      Transaction           `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	SubSubCategoryID uint64                `gorm:"not null" json:"sub_sub_category_id"`
	SubSubCategory   ServiceSubSubCategory `gorm:"foreignKey:SubSubCategoryID" json:"-"`
	ProjectDeadline  time.Time             `gorm:"not null" json:"project_deadline"`
	BriefDescription string                `gorm:"type:text;not null" json:"brief_description"`
}

type ServiceMilestone struct {
	ID             string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID  string              `gorm:"type:uuid;not null;index" json:"transaction_id"`
	ServiceTx      TransactionServices `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	MilestoneIndex int                 `gorm:"type:integer;not null" json:"milestone_index"`
	Title          string              `gorm:"type:varchar(255);not null" json:"title"`
	Amount         int64               `gorm:"type:bigint;not null" json:"amount"`
	Status         string              `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	ReleasedAt     *time.Time          `json:"released_at,omitempty"`
	CreatedAt      time.Time           `gorm:"default:now()" json:"created_at"`
}

type TransactionEvents struct {
	TransactionID       string              `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	Transaction         Transaction         `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	SubSubCategoryID    uint64              `gorm:"not null" json:"sub_sub_category_id"`
	SubSubCategory      EventSubSubCategory `gorm:"foreignKey:SubSubCategoryID" json:"-"`
	EventName           string              `gorm:"type:varchar(255);not null" json:"event_name"`
	EventStartTime      time.Time           `gorm:"not null" json:"event_start_time"`
	EventEndTime        time.Time           `gorm:"not null" json:"event_end_time"`
	TicketQuantityTotal int                 `gorm:"type:integer;not null;default:0" json:"ticket_quantity_total"`
}

type EventVendorPayout struct {
	ID                    string            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID         string            `gorm:"type:uuid;not null;index" json:"transaction_id"`
	EventTx               TransactionEvents `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	VendorUserID          *string           `gorm:"type:uuid" json:"vendor_user_id,omitempty"` // internal vendor (platform user) -> internal wallet credit; nil = external vendor -> Midtrans disbursement
	VendorName            string            `gorm:"type:varchar(255);not null" json:"vendor_name"`
	VendorBankName        string            `gorm:"type:varchar(100);not null" json:"vendor_bank_name"`
	VendorAccountNumber   string            `gorm:"type:varchar(100);not null" json:"vendor_account_number"`
	AmountRequested       int64             `gorm:"type:bigint;not null" json:"amount_requested"`
	ExpenseDescription    string            `gorm:"type:text;not null" json:"expense_description"`
	InvoiceFileURL        string            `gorm:"type:text;not null" json:"invoice_file_url"`
	PayoutPhase           string            `gorm:"type:varchar(100);not null;default:'FINAL_SETTLEMENT'" json:"payout_phase"`
	Status                string            `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	IsDisbursedByMidtrans bool              `gorm:"type:boolean;not null;default:false" json:"is_disbursed_by_midtrans"`
	DisbursedAt           *time.Time        `json:"disbursed_at,omitempty"`
	ReviewedBy            *string           `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Reviewer              *UserProfile      `gorm:"foreignKey:ReviewedBy"`
	ReviewedAt            *time.Time        `json:"reviewed_at,omitempty"`
	CreatedAt             time.Time         `gorm:"default:now()" json:"created_at"`
}

type EventVendorAllocation struct {
	ID               string            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID    string            `gorm:"type:uuid;not null;index" json:"transaction_id"`
	EventTx          TransactionEvents `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	VendorID         string            `gorm:"type:uuid;not null" json:"vendor_id"`
	Vendor           VendorProfile     `gorm:"foreignKey:VendorID"`
	AllocatedAmount  int64             `gorm:"type:bigint;not null" json:"allocated_amount"`
	ActualPaidAmount int64             `gorm:"type:bigint;default:0" json:"actual_paid_amount"`
	Status           string            `gorm:"type:varchar(50);default:'PLEDGED'" json:"status"`
	CreatedAt        time.Time         `gorm:"default:now()" json:"created_at"`
}

type EventOfficialDetails struct {
	TransactionID string            `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	EventTx       TransactionEvents `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	OrganizerID   string            `gorm:"type:uuid;not null" json:"organizer_id"`
	Organizer     VendorProfile     `gorm:"foreignKey:OrganizerID"`
	ManagementFee int64             `gorm:"type:bigint;not null" json:"management_fee"`
	ApprovedAt    time.Time         `gorm:"default:now()" json:"approved_at"`
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	GetTransactionByID(ctx context.Context, id string) (*Transaction, error)
	GetTransactionByMidtransOrderID(ctx context.Context, orderID string) (*Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id string, status TransactionStatus) error
	GetExpiredLockedTransactions(ctx context.Context) ([]string, error)

	// Detail-container write paths. The type-specific detail row (plus its
	// milestones / vendor allocations) must be created inside the same UnitOfWork
	// as the master `transactions` row so the container can never reference a
	// master that rolled back (or vice versa).
	CreateGoodsDetail(ctx context.Context, tg *TransactionGoods) error
	CreateServicesDetail(ctx context.Context, ts *TransactionServices, milestones []ServiceMilestone) error
	CreateEventsDetail(ctx context.Context, te *TransactionEvents, allocations []EventVendorAllocation) error

	// GetActiveEventVendorPayoutsTotal returns the sum of amount_requested over
	// all payouts of an event transaction that have not reached DISBURSED/
	// terminal state — used to cap new invoice submissions against the escrow.
	GetActiveEventVendorPayoutsTotal(ctx context.Context, txID string) (int64, error)

	// GetReleasedMilestonesTotalByTxID returns the sum of amounts of milestones
	// already RELEASED for a services transaction. Used to compute the escrow that
	// remains held when refunding a disputed transaction, so a refund never pays
	// out more than what is still locked.
	GetReleasedMilestonesTotalByTxID(ctx context.Context, transactionID string) (int64, error)

	GetMilestoneByID(ctx context.Context, id string) (*ServiceMilestone, error)
	UpdateMilestoneStatus(ctx context.Context, id string, status string) error
	GetEventVendorPayoutsByTxID(ctx context.Context, txID string) ([]EventVendorPayout, error)
	GetEventVendorPayoutByID(ctx context.Context, payoutID string) (*EventVendorPayout, error)
	UpdateEventVendorPayoutStatus(ctx context.Context, id string, status string) error
	// MarkEventVendorPayoutDisbursed records that an external vendor's bank payout
	// was completed out-of-band. Sets DISBURSED + the disbursing admin + timestamps.
	// Does not move money in-system (see ADR-0002).
	MarkEventVendorPayoutDisbursed(ctx context.Context, payoutID string, adminID string) error

	// UpdateBlockchainLog stores the on-chain audit-log hash & timestamp after
	// the relayer successfully records the transaction to Avalanche.
	UpdateBlockchainLog(ctx context.Context, txID string, txHash string) error

	// Read paths for the user-facing transaction APIs.
	ListTransactionsByUser(ctx context.Context, userID string, limit, offset int) ([]Transaction, error)
	GetGoodsDetailByTxID(ctx context.Context, txID string) (*TransactionGoods, error)
	GetServicesDetailByTxID(ctx context.Context, txID string) (*TransactionServices, error)
	GetMilestonesByTxID(ctx context.Context, txID string) ([]ServiceMilestone, error)
	GetEventsDetailByTxID(ctx context.Context, txID string) (*TransactionEvents, error)
}
