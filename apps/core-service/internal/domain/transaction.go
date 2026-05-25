package domain

import (
	"time"
)

// Mengunci aturan finansial Rekberkuy di level kode Go
const (
	MaxMemberEventLimit = 10000000 // Rp 10 Juta
	QRISTotalFeePercent = 0.01     // 1%
	WithdrawFeeToUser   = 7500     // Rp 7.500

	// Kunci Tambahan untuk Mengatasi Error di Usecase
	FeeGoodsRekberPay    = 2500 // Sesuai aturan 1 (Promo)
	FeeGoodsNonRekberPay = 5000 // Sesuai aturan 1 (Non-RekberPay)
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

// Transaction mewakili tabel master 'transactions'
type Transaction struct {
	ID                 string            `json:"id"`
	BuyerID            string            `json:"buyer_id"`
	SellerID           string            `json:"seller_id"`
	Type               RekberType        `json:"type"`
	Status             TransactionStatus `json:"status"`
	AmountBase         int64             `json:"amount_base"`
	ShippingFee        int64             `json:"shipping_fee"`
	ServiceFee         int64             `json:"service_fee"`
	MidtransFee        int64             `json:"midtrans_fee"`
	AmountGross        int64             `json:"amount_gross"`
	AmountNet          int64             `json:"amount_net"`
	MidtransOrderID    string            `json:"midtrans_order_id"`
	IdempotencyKey     string            `json:"idempotency_key"`
	PaymentMethod      string            `json:"payment_method"` // MIDTRANS atau REKBERPAY
	BlockchainTxHash   *string           `json:"blockchain_tx_hash,omitempty"`
	BlockchainLoggedAt *time.Time        `json:"blockchain_logged_at,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// TransactionGoods mewakili tabel 'transaction_goods' (Lini 1)
type TransactionGoods struct {
	TransactionID          string     `json:"transaction_id"`
	ShippingCourier        string     `json:"shipping_courier"`
	ShippingTrackingNumber *string    `json:"shipping_tracking_number,omitempty"`
	ShippingAddress        string     `json:"shipping_address"`
	AutoConfirmDeadline    time.Time  `json:"auto_confirm_deadline"`
}

// TransactionServices mewakili tabel 'transaction_services' (Lini 2)
type TransactionServices struct {
	TransactionID   string    `json:"transaction_id"`
	ProjectDeadline time.Time `json:"project_deadline"`
	BriefDescription string   `json:"brief_description"`
}

// ServiceMilestone mewakili tabel 'service_milestones' (Escrow 50:50)
type ServiceMilestone struct {
	ID            string     `json:"id"`
	TransactionID string     `json:"transaction_id"`
	MilestoneIndex int       `json:"milestone_index"`
	Title         string     `json:"title"`
	Amount        int64      `json:"amount"`
	Status        string     `json:"status"` // PENDING, RELEASED
	ReleasedAt    *time.Time `json:"released_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TransactionEvents mewakili tabel 'transaction_events' (Lini 3)
type TransactionEvents struct {
	TransactionID      string    `json:"transaction_id"`
	EventName          string    `json:"event_name"`
	EventStartTime     time.Time `json:"event_start_time"`
	EventEndTime       time.Time `json:"event_end_time"`
	TicketQuantityTotal int      `json:"ticket_quantity_total"`
}

// EventVendorPayout mewakili tabel 'event_vendor_payouts' (Split 30:70 ke Vendor)
type EventVendorPayout struct {
	ID                    string    `json:"id"`
	TransactionID         string    `json:"transaction_id"`
	VendorName            string    `json:"vendor_name"`
	VendorBankName        string    `json:"vendor_bank_name"`
	VendorAccountNumber   string    `json:"vendor_account_number"`
	AmountRequested       int64     `json:"amount_requested"`
	ExpenseDescription    string    `json:"expense_description"`
	InvoiceFileURL        string    `json:"invoice_file_url"`
	PayoutPhase           string    `json:"payout_phase"` // OPERATIONAL_DP, FINAL_SETTLEMENT
	Status                string    `json:"status"`       // PENDING, APPROVED, REJECTED
	IsDisbursedByMidtrans bool      `json:"is_disbursed_by_midtrans"`
	ReviewedBy            *string   `json:"reviewed_by,omitempty"`
	ReviewedAt            *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

// Dispute mewakili tabel resolusi sengketa 'disputes'
type Dispute struct {
	ID                string     `json:"id"`
	TransactionID     string     `json:"transaction_id"`
	RaisedBy          string     `json:"raised_by"`
	TargetPartyID     *string    `json:"target_party_id,omitempty"`
	MediatorID        *string    `json:"mediator_id,omitempty"`
	Reason            string     `json:"reason"`
	EvidenceURL       *string    `json:"evidence_url,omitempty"`
	IsResolved        bool       `json:"is_resolved"`
	ResolutionSummary *string    `json:"resolution_summary,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}