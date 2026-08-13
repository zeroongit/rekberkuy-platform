package handlers_test

// Endpoint handler tests. Handlers are thin wrappers (binding -> usecase -> JSON),
// so we cover (a) validation/auth failure paths with nil usecases (binding fails
// before any usecase call) and (b) success paths for representative endpoints
// using small inline mocks.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ---------------------------------------------------------------------------
// Minimal inline mocks (defaults are nil-safe / success).
// ---------------------------------------------------------------------------

type hTxRepo struct {
	domain.TransactionRepository
	onCreate func(ctx context.Context, tx *domain.Transaction) error
}

func (m *hTxRepo) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	if m.onCreate != nil {
		return m.onCreate(ctx, tx)
	}
	return nil
}

type hWalletRepo struct {
	domain.WalletRepository
	onUpdateBalance      func(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error
	onMarkWalletTxStatus func(ctx context.Context, orderID string, status domain.WalletTxStatus) error
}

func (m *hWalletRepo) UpdateBalanceTx(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error {
	if m.onUpdateBalance != nil {
		return m.onUpdateBalance(ctx, rec, mod)
	}
	return nil
}
func (m *hWalletRepo) MarkWalletTxStatusByOrderID(ctx context.Context, orderID string, status domain.WalletTxStatus) error {
	if m.onMarkWalletTxStatus != nil {
		return m.onMarkWalletTxStatus(ctx, orderID, status)
	}
	return nil
}

type hUserRepo struct {
	domain.UserRepository
	onGetProfileByEmail func(ctx context.Context, email string) (*domain.UserProfile, error)
}

func (m *hUserRepo) CreateProfile(ctx context.Context, u *domain.UserProfile) error { return nil }

func (m *hUserRepo) GetProfileByEmail(ctx context.Context, email string) (*domain.UserProfile, error) {
	if m.onGetProfileByEmail != nil {
		return m.onGetProfileByEmail(ctx, email)
	}
	return nil, errors.New("not found")
}

type hFinanceRepo struct{ domain.FinanceRepository }

func (m *hFinanceRepo) UpdatePlatformFinance(ctx context.Context, e, r, m2 int64) error { return nil }

type hUoW struct {
	domain.UnitOfWork
}

func (m *hUoW) Do(ctx context.Context, fn func(ctx context.Context, stores domain.TxStores) error) error {
	return fn(ctx, domain.TxStores{
		Users:        hUserStore{},
		Wallets:      hWalletStore{},
		Transactions: hTxStore{},
		Finance:      hFinStore{},
	})
}

// store-scoped stubs used inside UoW closure
type hUserStore struct{ domain.UserRepository }

func (hUserStore) CreateProfile(ctx context.Context, u *domain.UserProfile) error { return nil }

type hWalletStore struct{ domain.WalletRepository }

func (hWalletStore) UpdateBalanceTx(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error {
	return nil
}

func (hWalletStore) CreateWallet(ctx context.Context, userID string) error { return nil }

type hTxStore struct{ domain.TransactionRepository }

type hFinStore struct{ domain.FinanceRepository }

func (hFinStore) UpdatePlatformFinance(ctx context.Context, e, r, m2 int64) error { return nil }

type hMidtrans struct{ domain.MidtransClient }

func (m *hMidtrans) CreateSnapTransaction(ctx context.Context, req domain.SnapRequest) (domain.SnapResult, error) {
	return domain.SnapResult{Token: "tok", RedirectURL: "https://snap/x"}, nil
}

type hFraud struct{ domain.FraudClient }

func (m *hFraud) AnalyzeTransactionRisk(ctx context.Context, uid string, amt int64) (float64, bool, error) {
	return 0.1, true, nil
}

type hRelayer struct{ domain.Relayer }

func (m *hRelayer) LogTransactionOnChain(ctx context.Context, id string, amt int64, b, s string) (string, error) {
	return "0xmock", nil
}

type hVendorRepo struct{ domain.VendorRepository }

func (m *hVendorRepo) CreateVendor(ctx context.Context, v *domain.VendorProfile) error { return nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func init() { gin.SetMode(gin.TestMode) }

func doJSON(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// routerWithUser builds a gin engine where the route runs under a fake middleware
// that injects user_id (simulating AuthMiddleware.RequireRole success).
func routerWithUser(route string, h gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.POST(route, func(c *gin.Context) { c.Set("user_id", "user-uuid-1"); h(c) })
	return r
}

func routerNoUser(route string, h gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.POST(route, h)
	return r
}

// ---------------------------------------------------------------------------
// Wallet top-up handler
// ---------------------------------------------------------------------------

func TestCreateTopUpHandler_Success(t *testing.T) {
	walletRepo := &hWalletRepo{}
	uu := usecase.NewUserUsecase(&hUoW{}, &hUserRepo{}, walletRepo, &hMidtrans{})
	h := handlers.NewWalletHandler(uu)

	body, _ := json.Marshal(map[string]int64{"amount": 50000})
	w := doJSON(routerWithUser("/topup", h.CreateTopUpHandler), http.MethodPost, "/topup", string(body))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["snap_token"] != "tok" {
		t.Errorf("snap_token = %v, want tok", resp["snap_token"])
	}
}

func TestCreateTopUpHandler_InvalidAmount(t *testing.T) {
	h := handlers.NewWalletHandler(nil)
	body := `{"amount":0}`
	w := doJSON(routerWithUser("/topup", h.CreateTopUpHandler), http.MethodPost, "/topup", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreateTopUpHandler_Unauthenticated(t *testing.T) {
	h := handlers.NewWalletHandler(nil)
	body := `{"amount":5000}`
	w := doJSON(routerNoUser("/topup", h.CreateTopUpHandler), http.MethodPost, "/topup", body)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Auth register / login handlers
// ---------------------------------------------------------------------------

func newAuthHandler(userRepo *hUserRepo) *handlers.AuthHandler {
	token := usecase.NewTokenService("secret", time.Hour)
	au := usecase.NewAuthUsecase(&hUoW{}, userRepo, token)
	return handlers.NewAuthHandler(au)
}

func TestAuthRegisterHandler_Success(t *testing.T) {
	h := newAuthHandler(&hUserRepo{})
	body := `{"email":"jane@example.com","username":"jane","password":"password123","full_name":"Jane Doe"}`
	w := doJSON(routerNoUser("/register", h.RegisterHandler), http.MethodPost, "/register", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
}

func TestAuthRegisterHandler_InvalidPayload(t *testing.T) {
	h := newAuthHandler(&hUserRepo{})
	body := `{"email":"not-an-email","username":"x","password":"short","full_name":"X"}`
	w := doJSON(routerNoUser("/register", h.RegisterHandler), http.MethodPost, "/register", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAuthLoginHandler_InvalidPayload(t *testing.T) {
	h := newAuthHandler(&hUserRepo{})
	body := `{"email":""}` // missing required fields
	w := doJSON(routerNoUser("/login", h.LoginHandler), http.MethodPost, "/login", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAuthLoginHandler_BadCredentials(t *testing.T) {
	h := newAuthHandler(&hUserRepo{}) // default GetProfileByEmail => not found
	body := `{"email":"ghost@example.com","password":"whatever"}`
	w := doJSON(routerNoUser("/login", h.LoginHandler), http.MethodPost, "/login", body)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestGenerateTokenTestHandler_Success(t *testing.T) {
	h := handlers.NewUserHandler(nil, "test-secret")
	body := `{"user_id":"550e8400-e29b-41d4-a716-446655440000","username":"u","role":"USER"}`
	w := doJSON(routerNoUser("/token", h.GenerateTokenTestHandler), http.MethodPost, "/token", body)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["access_token"] == nil {
		t.Error("access_token must exist")
	}
}

// ---------------------------------------------------------------------------
// Goods lock/release handlers
// ---------------------------------------------------------------------------

func newGoodsHandler() *handlers.TransactionGoodsHandler {
	uu := usecase.NewTransactionGoodsUsecase(&hUoW{}, &hTxRepo{}, usecase.NewFinanceCalculator(), &hFraud{}, &hRelayer{})
	return handlers.NewTransactionGoodsHandler(uu)
}

func TestLockFundsGoodsHandler_Success(t *testing.T) {
	h := newGoodsHandler()
	body := `{"buyer_id":"550e8400-e29b-41d4-a716-446655440000","seller_id":"550e8400-e29b-41d4-a716-446655440001","amount_base":100000,"is_rekber_pay":true,"seller_tier":"BRONZE","shipping_fee":10000,"payment_method":"REKBERPAY","idempotency_key":"idem-1"}`
	w := doJSON(routerNoUser("/lock", h.LockFundsGoodsHandler), http.MethodPost, "/lock", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
}

func TestLockFundsGoodsHandler_InvalidPayload(t *testing.T) {
	h := newGoodsHandler()
	body := `{"amount_base":0}` // missing required fields
	w := doJSON(routerNoUser("/lock", h.LockFundsGoodsHandler), http.MethodPost, "/lock", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestReleaseGoodsHandler_InvalidPayload(t *testing.T) {
	h := newGoodsHandler()
	body := `{"transaction_id":"not-a-uuid"}`
	w := doJSON(routerNoUser("/release", h.ReleaseGoodsHandler), http.MethodPost, "/release", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------------------------------------------------------------------------
// KYC submit handler (auth + validation)
// ---------------------------------------------------------------------------

func TestSubmitKYCHandler_Unauthenticated(t *testing.T) {
	h := handlers.NewKYCHandler(nil)
	w := doJSON(routerNoUser("/kyc", h.SubmitKYCHandler), http.MethodPost, "/kyc", `{}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestSubmitKYCHandler_InvalidPayload(t *testing.T) {
	h := handlers.NewKYCHandler(nil)
	body := `{"target_role":"USER"}` // invalid enum + missing fields
	w := doJSON(routerWithUser("/kyc", h.SubmitKYCHandler), http.MethodPost, "/kyc", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Vendor register handler
// ---------------------------------------------------------------------------

func TestRegisterVendorHandler_Success(t *testing.T) {
	vu := usecase.NewVendorUsecase(&hVendorRepo{})
	h := handlers.NewVendorHandler(vu)
	body := `{"business_name":"Katering Sedap","category":"CATERING"}`
	w := doJSON(routerWithUser("/vendors", h.RegisterVendorHandler), http.MethodPost, "/vendors", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
}

func TestRegisterVendorHandler_Unauthenticated(t *testing.T) {
	h := handlers.NewVendorHandler(nil)
	w := doJSON(routerNoUser("/vendors", h.RegisterVendorHandler), http.MethodPost, "/vendors", `{}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRegisterVendorHandler_InvalidPayload(t *testing.T) {
	h := handlers.NewVendorHandler(nil)
	body := `{"business_name":""}` // missing required category + empty name
	w := doJSON(routerWithUser("/vendors", h.RegisterVendorHandler), http.MethodPost, "/vendors", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Services lock + Events lock (success + validation)
// ---------------------------------------------------------------------------

func newServicesHandler() *handlers.TransactionServicesHandler {
	uu := usecase.NewTransactionServicesUsecase(&hUoW{}, &hTxRepo{}, usecase.NewFinanceCalculator(), &hFraud{}, &hRelayer{})
	return handlers.NewTransactionServicesHandler(uu)
}

func newEventsHandler() *handlers.TransactionEventsHandler {
	uu := usecase.NewTransactionEventsUsecase(&hUoW{}, &hTxRepo{}, &hWalletRepo{}, usecase.NewFinanceCalculator(), &hFraud{}, &hRelayer{})
	return handlers.NewTransactionEventsHandler(uu)
}

const validLockBody = `{"buyer_id":"550e8400-e29b-41d4-a716-446655440000","seller_id":"550e8400-e29b-41d4-a716-446655440001","amount_base":100000,"is_rekber_pay":true,"seller_tier":"BRONZE","payment_method":"REKBERPAY","idempotency_key":"idem-x"}`

func TestLockFundsServicesHandler_Success(t *testing.T) {
	h := newServicesHandler()
	w := doJSON(routerNoUser("/slock", h.LockFundsServicesHandler), http.MethodPost, "/slock", validLockBody)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
}

func TestLockFundsServicesHandler_InvalidPayload(t *testing.T) {
	h := newServicesHandler()
	w := doJSON(routerNoUser("/slock", h.LockFundsServicesHandler), http.MethodPost, "/slock", `{"amount_base":0}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestReleaseMilestoneServicesHandler_InvalidPayload(t *testing.T) {
	h := newServicesHandler()
	w := doJSON(routerNoUser("/srel", h.ReleaseMilestoneHandler), http.MethodPost, "/srel", `{"milestone_id":"not-a-uuid"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLockFundsEventsHandler_Success(t *testing.T) {
	h := newEventsHandler()
	w := doJSON(routerNoUser("/elock", h.LockFundsEventsHandler), http.MethodPost, "/elock", validLockBody)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
}

func TestLockFundsEventsHandler_InvalidPayload(t *testing.T) {
	h := newEventsHandler()
	w := doJSON(routerNoUser("/elock", h.LockFundsEventsHandler), http.MethodPost, "/elock", `{"amount_base":0}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestProcessEventVendorPayoutHandler_InvalidPayload(t *testing.T) {
	h := newEventsHandler()
	w := doJSON(routerNoUser("/epay", h.ProcessEventVendorPayoutHandler), http.MethodPost, "/epay", `{"transaction_id":"bad"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestReleaseEventMilestoneHandler_InvalidPayload(t *testing.T) {
	h := newEventsHandler()
	w := doJSON(routerNoUser("/erel", h.ReleaseEventMilestoneHandler), http.MethodPost, "/erel", `{"payout_id":"bad"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------------------------------------------------------------------------
// CORS middleware smoke test
// ---------------------------------------------------------------------------

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(handlers.CORSMiddleware(nil))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	req, _ := http.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS Allow-Origin header must be set")
	}
}
