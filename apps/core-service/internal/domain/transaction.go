package domain

import (
	"time"
	"context"
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
	TransactionID          string      `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	Transaction            Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	ShippingCourier        string      `gorm:"type:varchar(100);not null" json:"shipping_courier"`
	ShippingTrackingNumber *string     `gorm:"type:varchar(255)" json:"shipping_tracking_number,omitempty"`
	ShippingAddress        string      `gorm:"type:text;not null" json:"shipping_address"`
	AutoConfirmDeadline    time.Time   `gorm:"not null" json:"auto_confirm_deadline"`
}

type TransactionServices struct {
	TransactionID    string      `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	Transaction      Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	ProjectDeadline  time.Time   `gorm:"not null" json:"project_deadline"`
	BriefDescription string      `gorm:"type:text" json:"brief_description"`
}

type ServiceMilestone struct {
	ID            string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID string              `gorm:"type:uuid;not null;index" json:"transaction_id"`
	ServiceTx     TransactionServices `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	MilestoneIndex int                 `gorm:"type:integer;not null" json:"milestone_index"`
	Title         string              `gorm:"type:varchar(255);not null" json:"title"`
	Amount        int64               `gorm:"type:bigint;not null" json:"amount"`
	Status        string              `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	ReleasedAt    *time.Time          `json:"released_at,omitempty"`
	CreatedAt     time.Time           `gorm:"default:now()" json:"created_at"`
}

type TransactionEvents struct {
	TransactionID       string      `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	Transaction         Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	EventName           string      `gorm:"type:varchar(255);not null" json:"event_name"`
	EventStartTime      time.Time   `gorm:"not null" json:"event_start_time"`
	EventEndTime        time.Time   `gorm:"not null" json:"event_end_time"`
	TicketQuantityTotal int         `gorm:"type:integer;not null;default:0" json:"ticket_quantity_total"`
}

type EventVendorPayout struct {
	ID                    string            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID         string            `gorm:"type:uuid;not null;index" json:"transaction_id"`
	EventTx               TransactionEvents `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	VendorName            string            `gorm:"type:varchar(255);not null" json:"vendor_name"`
	VendorBankName        string            `gorm:"type:varchar(100);not null" json:"vendor_bank_name"`
	VendorAccountNumber   string            `gorm:"type:varchar(100);not null" json:"vendor_account_number"`
	AmountRequested       int64             `gorm:"type:bigint;not null" json:"amount_requested"`
	ExpenseDescription    string            `gorm:"type:text;not null" json:"expense_description"`
	InvoiceFileURL        string            `gorm:"type:text;not null" json:"invoice_file_url"`
	PayoutPhase           string            `gorm:"type:varchar(100);not null;default:'FINAL_SETTLEMENT'" json:"payout_phase"`
	Status                string            `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	IsDisbursedByMidtrans bool              `gorm:"type:boolean;not null;default:false" json:"is_disbursed_by_midtrans"`
	ReviewedBy            *string           `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Reviewer              *UserProfile      `gorm:"foreignKey:ReviewedBy"`
	ReviewedAt            *time.Time        `json:"reviewed_at,omitempty"`
	CreatedAt             time.Time         `gorm:"default:now()" json:"created_at"`
}

type Dispute struct {
	ID                string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID     string      `gorm:"type:uuid;not null;unique" json:"transaction_id"`
	Transaction       Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	RaisedBy          string      `gorm:"type:uuid;not null" json:"raised_by"`
	TargetPartyID     *string     `gorm:"type:uuid" json:"target_party_id,omitempty"`
	MediatorID        *string     `gorm:"type:uuid" json:"mediator_id,omitempty"`
	Raiser            UserProfile `gorm:"foreignKey:RaisedBy"`
	Target            UserProfile `gorm:"foreignKey:TargetPartyID"`
	Mediator          UserProfile `gorm:"foreignKey:MediatorID"`
	Reason            string      `gorm:"type:text;not null" json:"reason"`
	EvidenceURL       *string     `gorm:"type:text" json:"evidence_url,omitempty"`
	IsResolved        bool        `gorm:"type:boolean;not null;default:false" json:"is_resolved"`
	ResolutionSummary *string     `gorm:"type:text" json:"resolution_summary,omitempty"`
	CreatedAt         time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time   `gorm:"default:now()" json:"updated_at"`
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	GetTransactionByID(ctx context.Context, id string) (*Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id string, status TransactionStatus) error
}