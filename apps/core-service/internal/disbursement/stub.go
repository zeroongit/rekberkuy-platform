package disbursement

import (
	"context"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

// This package implements the domain.DisbursementClient port.
//
// STATUS: stub only. The bank transfer for wallet withdrawals still completes
// OUT-OF-BAND (ADR-0002 stance); this stub serves the ESTIMATE half of the
// true-up model — it returns the configured MIDTRANS_DISBURSEMENT_FEE so the
// ledger books a sensible expected cost at request time, corrected to the real
// cost when the admin confirms the disbursement.
//
// When Midtrans Payout/Disbursement credentials arrive, add an HTTP adapter in
// this package (NewPayoutHTTPClient) calling the provider API: the actual fee
// from the API response becomes the source of truth, and the withdrawal flow
// can optionally execute transfers in-system instead of out-of-band.
type disbursementStub struct {
	estimate int64
}

// NewDisbursementStub returns a DisbursementClient reporting the configured
// flat estimate (MIDTRANS_DISBURSEMENT_FEE) for every transfer.
func NewDisbursementStub(estimate int64) domain.DisbursementClient {
	if estimate < 0 {
		estimate = 0
	}
	return &disbursementStub{estimate: estimate}
}

func (s *disbursementStub) ExecuteDisbursement(ctx context.Context, req domain.DisbursementRequest) (domain.DisbursementResult, error) {
	// No production caller yet — executing out-of-band is the current stance.
	// Kept honest: the stub refuses rather than pretending a transfer happened.
	return domain.DisbursementResult{}, fmt.Errorf("disbursement execution not wired (out-of-band stance, see ADR-0002)")
}

func (s *disbursementStub) EstimateDisbursementFee(ctx context.Context, req domain.DisbursementRequest) (int64, error) {
	return s.estimate, nil
}
