package domain

import (
	"context"
	"time"
)

type PlatformFinance struct {
	ID                 string    `json:"id"`
	TotalEscrowBalance int64     `json:"total_escrow_balance"` // Uang pembeli yang sedang dikunci
	TotalRevenue       int64     `json:"total_revenue"`        // Pendapatan bersih dari service fee
	TotalMidtransFees  int64     `json:"total_midtrans_fees"`   // Total biaya yang dibayarkan ke Midtrans
	UpdatedAt          time.Time `json:"updated_at"`
}

type EventAuditResult struct {
	PlatformFee     int64 `json:"platform_fee"`
	AmountToVendor  int64 `json:"amount_to_vendor"`
	BonusToEO       int64 `json:"bonus_to_eo"`
	RefundToPeserta int64 `json:"refund_to_peserta"`
}

type FinanceRepository interface {
	GetPlatformFinance(ctx context.Context) (*PlatformFinance, error)
	UpdatePlatformFinance(ctx context.Context, escrowDelta int64, revenueDelta int64, midtransFeeDelta int64) error
}
