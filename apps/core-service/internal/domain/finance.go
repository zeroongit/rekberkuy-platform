package domain

import (
	"context"
	"time"
)

type PlatformFinance struct {
	ID                 string    `json:"id"`
	TotalEscrowBalance int64     `json:"total_escrow_balance"` // Buyer funds currently locked
	TotalRevenue       int64     `json:"total_revenue"`        // Net revenue from service fee
	TotalMidtransFees  int64     `json:"total_midtrans_fees"`  // Total fees paid to Midtrans
	UpdatedAt          time.Time `json:"updated_at"`
}

type EventAuditResult struct {
	PlatformFee     int64 `json:"platform_fee"`
	AmountToVendor  int64 `json:"amount_to_vendor"`
	BonusToEO       int64 `json:"bonus_to_eo"`
	RefundToPeserta int64 `json:"refund_to_peserta"`
}

// Platform fee & withdrawal cost constants. All money math flows through here
// so the escrow-capping and the settlement paths can never drift apart.
const (
	// EventPlatformFeePercent is the platform's cut of a settled event escrow,
	// taken from AmountGross before vendor payouts.
	EventPlatformFeePercent = 5
	// WithdrawFeeMidtransCost is the Midtrans disbursement cost charged to the
	// platform for executing the bank transfer. The user-facing
	// WithdrawFeeToUser (see transaction.go) is a GROSS fee: it already covers
	// this cost, and the remainder is the platform's net profit.
	WithdrawFeeMidtransCost = 4000
)

// EventPlatformFee computes the platform's cut of an event escrow settlement.
// Both the vendor-invoice cap and the settlement ledger must use this helper
// so the two paths stay arithmetically identical.
func EventPlatformFee(amountGross int64) int64 {
	return amountGross * EventPlatformFeePercent / 100
}

// WithdrawalNetProfit computes the platform's net profit on a withdrawal:
// the gross user-facing fee minus the Midtrans disbursement cost.
func WithdrawalNetProfit(grossFee int64) int64 {
	return grossFee - WithdrawFeeMidtransCost
}

type FinanceRepository interface {
	GetPlatformFinance(ctx context.Context) (*PlatformFinance, error)
	UpdatePlatformFinance(ctx context.Context, escrowDelta int64, revenueDelta int64, midtransFeeDelta int64) error
}
