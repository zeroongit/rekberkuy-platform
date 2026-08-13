package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// txReviewRepo returns a RELEASED transaction whose buyer matches the
// routerWithUser-injected user_id ("user-uuid-1").
type txReviewRepo struct {
	domain.TransactionRepository
	tx *domain.Transaction
}

func (m *txReviewRepo) GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error) {
	if m.tx != nil {
		return m.tx, nil
	}
	return nil, errors.New("not found")
}

type reviewCreateRepo struct {
	domain.ReviewRepository
	last *domain.Review
	list []domain.Review
}

func (m *reviewCreateRepo) CreateReview(ctx context.Context, rv *domain.Review) error {
	m.last = rv
	return nil
}
func (m *reviewCreateRepo) GetReviewsByReviewee(ctx context.Context, id string) ([]domain.Review, error) {
	return m.list, nil
}
func (m *reviewCreateRepo) GetAverageRatingForUser(ctx context.Context, userID string) (float64, error) {
	return 0, nil
}

func newReviewHandler() (*handlers.ReviewHandler, *reviewCreateRepo) {
	txr := &txReviewRepo{tx: &domain.Transaction{ID: "tx-1", BuyerID: "user-uuid-1", SellerID: "seller-1", Status: domain.StatusReleased}}
	rr := &reviewCreateRepo{list: []domain.Review{{ID: "r1", Rating: 4}}}
	uu := usecase.NewReviewUsecase(txr, rr)
	return handlers.NewReviewHandler(uu), rr
}

func TestCreateReviewHandler_Success(t *testing.T) {
	h, rr := newReviewHandler()
	body := `{"transaction_id":"550e8400-e29b-41d4-a716-446655440000","rating":5,"comment":"ok"}`
	// routerWithUser injects user_id="user-uuid-1", which matches the tx buyer.
	w := doJSON(routerWithUser("/reviews", h.CreateReviewHandler), http.MethodPost, "/reviews", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d, want 201; body=%s", w.Code, w.Body.String())
	}
	if rr.last == nil || rr.last.Rating != 5 {
		t.Errorf("review not created as expected: %+v", rr.last)
	}
}

func TestCreateReviewHandler_InvalidPayload(t *testing.T) {
	h, _ := newReviewHandler()
	body := `{"transaction_id":"not-a-uuid","rating":9}`
	w := doJSON(routerWithUser("/reviews", h.CreateReviewHandler), http.MethodPost, "/reviews", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400", w.Code)
	}
}

func TestCreateReviewHandler_Unauthenticated(t *testing.T) {
	h, _ := newReviewHandler()
	body := `{"transaction_id":"550e8400-e29b-41d4-a716-446655440000","rating":5}`
	w := doJSON(routerNoUser("/reviews", h.CreateReviewHandler), http.MethodPost, "/reviews", body)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
}

func TestGetReviewsForUserHandler_Success(t *testing.T) {
	h, _ := newReviewHandler()
	r := gin.New()
	r.GET("/users/:id/reviews", h.GetReviewsForUserHandler)

	req, _ := http.NewRequest(http.MethodGet, "/users/seller-1/reviews", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["reviews"] == nil {
		t.Error("expected a reviews field in the response")
	}
}
