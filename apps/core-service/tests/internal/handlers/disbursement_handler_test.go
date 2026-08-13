package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type disburseTxRepo struct {
	domain.TransactionRepository
	payout *domain.EventVendorPayout
	marked bool
}

func (m *disburseTxRepo) GetEventVendorPayoutByID(ctx context.Context, id string) (*domain.EventVendorPayout, error) {
	if m.payout != nil {
		return m.payout, nil
	}
	return nil, errors.New("not found")
}

func (m *disburseTxRepo) MarkEventVendorPayoutDisbursed(ctx context.Context, payoutID, adminID string) error {
	m.marked = true
	return nil
}

func newDisbursementHandler() (*handlers.DisbursementHandler, *disburseTxRepo) {
	txr := &disburseTxRepo{payout: &domain.EventVendorPayout{ID: "pay-1", Status: domain.VendorPayoutPendingDisbursement}}
	uu := usecase.NewDisbursementUsecase(txr)
	return handlers.NewDisbursementHandler(uu), txr
}

// routerWithAdmin injects user_id (the Admin actor) for a param-bearing route.
func routerWithAdmin(route string, h gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.POST(route, func(c *gin.Context) { c.Set("user_id", "admin-1"); h(c) })
	return r
}

func TestMarkDisbursedHandler_Success(t *testing.T) {
	h, txr := newDisbursementHandler()
	r := routerWithAdmin("/payouts/:id/disburse", h.MarkDisbursedHandler)
	w := doJSON(r, http.MethodPost, "/payouts/pay-1/disburse", `{}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", w.Code, w.Body.String())
	}
	if !txr.marked {
		t.Error("payout was not marked disbursed")
	}
}

func TestMarkDisbursedHandler_Unauthenticated(t *testing.T) {
	h, _ := newDisbursementHandler()
	r := gin.New()
	r.POST("/payouts/:id/disburse", h.MarkDisbursedHandler)
	w := doJSON(r, http.MethodPost, "/payouts/pay-1/disburse", `{}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
}

func TestMarkDisbursedHandler_WrongStatus(t *testing.T) {
	txr := &disburseTxRepo{payout: &domain.EventVendorPayout{ID: "pay-1", Status: domain.VendorPayoutApproved}}
	uu := usecase.NewDisbursementUsecase(txr)
	h := handlers.NewDisbursementHandler(uu)
	r := routerWithAdmin("/payouts/:id/disburse", h.MarkDisbursedHandler)
	w := doJSON(r, http.MethodPost, "/payouts/pay-1/disburse", `{}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400 (wrong status)", w.Code)
	}
}
