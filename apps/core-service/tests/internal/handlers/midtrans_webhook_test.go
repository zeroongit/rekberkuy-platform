package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
)

// mockMidtrans for handler tests.
type mockMidtrans struct {
	verify bool
}

func (m *mockMidtrans) CreateSnapTransaction(ctx context.Context, req domain.SnapRequest) (domain.SnapResult, error) {
	return domain.SnapResult{}, nil
}
func (m *mockMidtrans) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	return m.verify
}

func newWebhookRouter(h *handlers.MidtransWebhookHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/hook", h.NotificationHandler)
	return r
}

func TestMidtransWebhook_NilMidtransReturns503(t *testing.T) {
	h := handlers.NewMidtransWebhookHandler(nil, nil, nil, nil, nil)
	r := newWebhookRouter(h)

	body, _ := json.Marshal(domain.MidtransNotification{OrderID: "x"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503 (midtrans not configured)", w.Code)
	}
}

func TestMidtransWebhook_BadPayloadReturns400(t *testing.T) {
	h := handlers.NewMidtransWebhookHandler(&mockMidtrans{verify: true}, nil, nil, nil, nil)
	r := newWebhookRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/hook", bytes.NewReader([]byte("{not json")))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestMidtransWebhook_InvalidSignatureReturns403(t *testing.T) {
	h := handlers.NewMidtransWebhookHandler(&mockMidtrans{verify: false}, nil, nil, nil, nil)
	r := newWebhookRouter(h)

	body, _ := json.Marshal(domain.MidtransNotification{OrderID: "REKBERKUY-TOPUP-x", StatusCode: "200", GrossAmount: "100.00", SignatureKey: "bad"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403 (invalid signature)", w.Code)
	}
}
