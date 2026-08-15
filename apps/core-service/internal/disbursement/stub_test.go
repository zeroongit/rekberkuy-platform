package disbursement

import (
	"context"
	"testing"

	"rekberkuy/core-service/internal/domain"
)

func TestStub_EstimateReturnsConfiguredFee(t *testing.T) {
	c := NewDisbursementStub(4000)
	fee, err := c.EstimateDisbursementFee(context.Background(), domain.DisbursementRequest{Amount: 500000})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if fee != 4000 {
		t.Errorf("estimate = %d, want 4000", fee)
	}
}

func TestStub_ClampsNegativeEstimate(t *testing.T) {
	c := NewDisbursementStub(-5)
	fee, err := c.EstimateDisbursementFee(context.Background(), domain.DisbursementRequest{})
	if err != nil || fee != 0 {
		t.Errorf("negative config must clamp to 0, got %d err %v", fee, err)
	}
}

func TestStub_ExecuteRefuses(t *testing.T) {
	// The stub must never pretend a bank transfer happened — the honest
	// refusal documents the out-of-band stance until the HTTP adapter lands.
	c := NewDisbursementStub(4000)
	if _, err := c.ExecuteDisbursement(context.Background(), domain.DisbursementRequest{}); err == nil {
		t.Fatal("stub ExecuteDisbursement must refuse instead of faking a transfer")
	}
}
