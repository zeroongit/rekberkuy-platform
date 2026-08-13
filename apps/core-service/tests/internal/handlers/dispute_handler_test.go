package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ---- Stubs that exercise the real DisputeUsecase through its UoW ------------

type dTxRepo struct {
	domain.TransactionRepository
	tx           *domain.Transaction
	frozenStatus domain.TransactionStatus
}

func (m *dTxRepo) GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error) {
	return m.tx, nil
}
func (m *dTxRepo) UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	m.frozenStatus = status
	return nil
}

type dWalletRepo struct {
	domain.WalletRepository
	lastWallet string
	lastMod    int64
}

func (m *dWalletRepo) UpdateBalanceTx(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error {
	m.lastWallet, m.lastMod = rec.WalletID, mod
	return nil
}

type dFinRepo struct{ domain.FinanceRepository }

func (dFinRepo) UpdatePlatformFinance(ctx context.Context, e, r, m2 int64) error { return nil }

type dDisputeRepo struct {
	domain.DisputeRepository
	byID    *domain.Dispute
	created *domain.Dispute
	ack     bool
	got     string
}

func (m *dDisputeRepo) CreateDispute(ctx context.Context, d *domain.Dispute) error {
	m.created = d
	return nil
}
func (m *dDisputeRepo) GetDisputeByID(ctx context.Context, id string) (*domain.Dispute, error) {
	m.got = id
	return m.byID, nil
}
func (m *dDisputeRepo) AcknowledgeDispute(ctx context.Context, id string) error {
	m.ack = true
	return nil
}
func (m *dDisputeRepo) ResolveDispute(ctx context.Context, id, adminID string, o domain.DisputeOutcome, s string) error {
	return nil
}

type dUoW struct{ stores domain.TxStores }

func (m *dUoW) Do(ctx context.Context, fn func(ctx context.Context, stores domain.TxStores) error) error {
	return fn(ctx, m.stores)
}

func newDisputeHandler() (*handlers.DisputeHandler, *dDisputeRepo, *dTxRepo) {
	txr := &dTxRepo{tx: &domain.Transaction{ID: "tx-1", BuyerID: "user-uuid-1", SellerID: "seller-1", Status: domain.StatusFundsLocked, AmountGross: 200000, AmountNet: 180000}}
	dr := &dDisputeRepo{byID: &domain.Dispute{ID: "d-1", TransactionID: "tx-1", Status: domain.DisputeStatusUnderReview}}
	uow := &dUoW{stores: domain.TxStores{Transactions: txr, Wallets: &dWalletRepo{}, Finance: dFinRepo{}, Disputes: dr}}
	uu := usecase.NewDisputeUsecase(uow, dr)
	return handlers.NewDisputeHandler(uu), dr, txr
}

func TestOpenDisputeHandler_Success(t *testing.T) {
	h, dr, _ := newDisputeHandler()
	body := `{"transaction_id":"550e8400-e29b-41d4-a716-446655440000","reason":"item not as described"}`
	w := doJSON(routerWithUser("/disputes", h.OpenDisputeHandler), http.MethodPost, "/disputes", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d, want 201; body=%s", w.Code, w.Body.String())
	}
	if dr.created == nil || dr.created.Status != domain.DisputeStatusOpen {
		t.Errorf("dispute not created as OPEN: %+v", dr.created)
	}
}

func TestOpenDisputeHandler_InvalidPayload(t *testing.T) {
	h, _, _ := newDisputeHandler()
	w := doJSON(routerWithUser("/disputes", h.OpenDisputeHandler), http.MethodPost, "/disputes", `{"reason":"no tx"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400", w.Code)
	}
}

func TestOpenDisputeHandler_Unauthenticated(t *testing.T) {
	h, _, _ := newDisputeHandler()
	w := doJSON(routerNoUser("/disputes", h.OpenDisputeHandler), http.MethodPost, "/disputes", `{"transaction_id":"550e8400-e29b-41d4-a716-446655440000","reason":"x"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
}

func TestAcknowledgeDisputeHandler_Success(t *testing.T) {
	h, dr, _ := newDisputeHandler()
	dr.byID.Status = domain.DisputeStatusOpen // acknowledge requires OPEN
	r := routerWithAdmin("/disputes/:id/acknowledge", h.AcknowledgeDisputeHandler)
	w := doJSON(r, http.MethodPost, "/disputes/d-1/acknowledge", `{}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", w.Code, w.Body.String())
	}
	if !dr.ack {
		t.Error("dispute was not acknowledged")
	}
}

func TestResolveDisputeHandler_Success(t *testing.T) {
	h, _, txr := newDisputeHandler()
	txr.tx.Status = domain.StatusDisputed // resolve requires the transaction to be DISPUTED
	r := routerWithAdmin("/disputes/:id/resolve", h.ResolveDisputeHandler)
	body := `{"outcome":"REFUND_BUYER","summary":"buyer was right"}`
	w := doJSON(r, http.MethodPost, "/disputes/d-1/resolve", body)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestResolveDisputeHandler_InvalidOutcome(t *testing.T) {
	h, _, _ := newDisputeHandler()
	r := routerWithAdmin("/disputes/:id/resolve", h.ResolveDisputeHandler)
	body := `{"outcome":"NOT_A_VALID_OUTCOME","summary":"x"}`
	w := doJSON(r, http.MethodPost, "/disputes/d-1/resolve", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400 for invalid outcome", w.Code)
	}
}

func TestGetDisputeHandler_Success(t *testing.T) {
	h, dr, _ := newDisputeHandler()
	r := gin.New()
	r.GET("/disputes/:id", h.GetDisputeHandler)

	req, _ := http.NewRequest(http.MethodGet, "/disputes/d-1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if dr.got != "d-1" {
		t.Errorf("queried dispute id=%s, want d-1", dr.got)
	}
}
