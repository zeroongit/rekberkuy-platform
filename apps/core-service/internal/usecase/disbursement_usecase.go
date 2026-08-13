package usecase

import (
	"context"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

// DisbursementUsecase closes the external-vendor payout lifecycle. Per ADR-0002,
// the bank transfer happens out-of-band; the Admin's mark-disbursed call only
// records completion — it moves no money in-system.
type DisbursementUsecase struct {
	txRepo domain.TransactionRepository
}

func NewDisbursementUsecase(txRepo domain.TransactionRepository) *DisbursementUsecase {
	return &DisbursementUsecase{txRepo: txRepo}
}

// MarkExternalPayoutDisbursed marks an external vendor payout DISBURSED.
// Refused unless the payout is currently PENDING_DISBURSEMENT (guards against
// double-marking and against marking internal-vendor payouts).
func (u *DisbursementUsecase) MarkExternalPayoutDisbursed(ctx context.Context, payoutID, adminID string) error {
	payout, err := u.txRepo.GetEventVendorPayoutByID(ctx, payoutID)
	if err != nil {
		return fmt.Errorf("failed to load vendor payout: %w", err)
	}
	if payout.Status != domain.VendorPayoutPendingDisbursement {
		return fmt.Errorf("only PENDING_DISBURSEMENT payouts can be marked disbursed (current: %s)", payout.Status)
	}
	if err := u.txRepo.MarkEventVendorPayoutDisbursed(ctx, payoutID, adminID); err != nil {
		return fmt.Errorf("failed to mark payout disbursed: %w", err)
	}
	return nil
}
