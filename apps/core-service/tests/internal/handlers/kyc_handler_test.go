package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ---- Stubs that exercise the real KYCUsecase through its UoW ----------------

type kRepo struct {
	domain.KYCRepository
	byID     *domain.KYCSubmission
	reviewed *domain.KYCSubmission
	pending  []domain.KYCSubmission
	aiSaved  bool
}

func (m *kRepo) SubmitKYC(ctx context.Context, k *domain.KYCSubmission) error { return nil }
func (m *kRepo) GetKYCByID(ctx context.Context, id string) (*domain.KYCSubmission, error) {
	return m.byID, nil
}
func (m *kRepo) ListPendingKYCs(ctx context.Context) ([]domain.KYCSubmission, error) {
	return m.pending, nil
}
func (m *kRepo) SaveKYCAIResult(ctx context.Context, userID string, score float64, reason string) error {
	m.aiSaved = true
	return nil
}
func (m *kRepo) ReviewKYC(ctx context.Context, id string, status domain.KYCStatus, adminID string, notes string) error {
	return nil
}

type kUserRepo struct {
	domain.UserRepository
	promotedRole domain.UserRole
	promotedID   string
}

func (m *kUserRepo) UpdateUserRole(ctx context.Context, id string, role domain.UserRole) error {
	m.promotedID, m.promotedRole = id, role
	return nil
}

type kUoW struct{ stores domain.TxStores }

func (m *kUoW) Do(ctx context.Context, fn func(ctx context.Context, stores domain.TxStores) error) error {
	return fn(ctx, m.stores)
}

type kClient struct {
	domain.KYCClient
}

func (kClient) VerifyIdentity(ctx context.Context, userID string, idCardURL string, selfieURL string, targetRole domain.UserRole) (float64, string, error) {
	return 0.9, "reference-ok", nil
}

func newKYCTestRouter(repo *kRepo, userRepo *kUserRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	uow := &kUoW{stores: domain.TxStores{KYC: repo, Users: userRepo}}
	u := usecase.NewKYCUsecase(repo, uow, kClient{})
	h := handlers.NewKYCHandler(u)

	r := gin.New()
	r.POST("/api/v1/kyc/submit", func(c *gin.Context) {
		c.Set("user_id", "user-uuid-1")
		h.SubmitKYCHandler(c)
	})
	r.GET("/api/v1/admin/kyc/pending", h.GetPendingKYCsHandler)
	r.GET("/api/v1/admin/kyc/:id", h.GetKYCDetailHandler)
	r.POST("/api/v1/admin/kyc/:id/review", func(c *gin.Context) {
		c.Set("user_id", "admin-uuid-1")
		h.ReviewKYCHandler(c)
	})
	return r
}

func TestKYCHandler_SubmitStoresAIReference(t *testing.T) {
	repo := &kRepo{}
	r := newKYCTestRouter(repo, &kUserRepo{})

	body := `{"target_role":"VERIFIED_MERCHANT","id_card_number":"3171010101900001","id_card_url":"https://cdn/ktp.jpg","selfie_url":"https://cdn/selfie.jpg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/kyc/submit", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}
	if !repo.aiSaved {
		t.Error("AI reference should be persisted on submit")
	}
}

func TestKYCHandler_SubmitInvalidPayload(t *testing.T) {
	r := newKYCTestRouter(&kRepo{}, &kUserRepo{})

	body := `{"target_role":"ADMIN","id_card_number":"123","id_card_url":"not-a-url","selfie_url":"also-not"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/kyc/submit", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (oneof role + url validation)", w.Code)
	}
}

func TestKYCHandler_ReviewApprove(t *testing.T) {
	repo := &kRepo{byID: &domain.KYCSubmission{
		ID: "kyc-1", UserID: "user-uuid-1", TargetRole: domain.RoleEventOrganizer, Status: domain.KYCPending,
	}}
	userRepo := &kUserRepo{}
	r := newKYCTestRouter(repo, userRepo)

	body := `{"approve":true,"notes":"documents genuine"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/kyc/kyc-1/review", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	if userRepo.promotedRole != domain.RoleEventOrganizer {
		t.Errorf("promoted role = %s, want EVENT_ORGANIZER", userRepo.promotedRole)
	}
	if userRepo.promotedID != "user-uuid-1" {
		t.Errorf("promoted user = %s, want user-uuid-1", userRepo.promotedID)
	}
}

func TestKYCHandler_ReviewAlreadyReviewedConflict(t *testing.T) {
	repo := &kRepo{byID: &domain.KYCSubmission{ID: "kyc-1", Status: domain.KYCApproved}}
	r := newKYCTestRouter(repo, &kUserRepo{})

	body := `{"approve":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/kyc/kyc-1/review", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}

func TestKYCHandler_ListPending(t *testing.T) {
	repo := &kRepo{pending: []domain.KYCSubmission{{ID: "kyc-1", Status: domain.KYCPending}}}
	r := newKYCTestRouter(repo, &kUserRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/kyc/pending", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestKYCHandler_Detail(t *testing.T) {
	repo := &kRepo{byID: &domain.KYCSubmission{ID: "kyc-1", Status: domain.KYCPending}}
	r := newKYCTestRouter(repo, &kUserRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/kyc/kyc-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}
