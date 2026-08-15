package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ---- Inline stubs exercising the real usecases ------------------------------

type wWalletRepo struct {
	domain.WalletRepository
	debitFails bool
	withdrawal *domain.WithdrawalRequest
}

func (m *wWalletRepo) UpdateBalanceTx(ctx context.Context, rec *domain.RekberPayTransaction, mod int64) error {
	if m.debitFails {
		return context.DeadlineExceeded
	}
	return nil
}
func (m *wWalletRepo) CreateWithdrawalRequest(ctx context.Context, w *domain.WithdrawalRequest) error {
	m.withdrawal = w
	return nil
}
func (m *wWalletRepo) GetBalance(ctx context.Context, userID string) (*domain.RekberPayWallet, error) {
	return &domain.RekberPayWallet{UserID: userID, Balance: 1000000}, nil
}
func (m *wWalletRepo) GetWalletTxHistory(ctx context.Context, userID string, limit, offset int) ([]domain.RekberPayTransaction, error) {
	return []domain.RekberPayTransaction{{ID: "wtx-1", WalletID: userID}}, nil
}
func (m *wWalletRepo) ListWithdrawalsByUser(ctx context.Context, userID string, limit, offset int) ([]domain.WithdrawalRequest, error) {
	return nil, nil
}

type wFinRepo struct{ domain.FinanceRepository }

func (wFinRepo) UpdatePlatformFinance(ctx context.Context, e, r, m2 int64) error { return nil }

type wUoW struct{ stores domain.TxStores }

func (m *wUoW) Do(ctx context.Context, fn func(ctx context.Context, stores domain.TxStores) error) error {
	return fn(ctx, m.stores)
}

type wCatRepo struct{ domain.CategoryRepository }

func (wCatRepo) ListGoodsCategories(ctx context.Context) ([]domain.GoodsCategory, error) {
	return []domain.GoodsCategory{{ID: 1, Name: "Electronics"}}, nil
}
func (wCatRepo) ListServiceCategories(ctx context.Context) ([]domain.ServiceCategory, error) {
	return nil, nil
}
func (wCatRepo) ListEventCategories(ctx context.Context) ([]domain.EventCategory, error) {
	return nil, nil
}
func (wCatRepo) ListVendorSubCategories(ctx context.Context) ([]domain.VendorSubCategory, error) {
	return nil, nil
}

type wVendorRepo struct{ domain.VendorRepository }

func (wVendorRepo) CreateVendor(ctx context.Context, v *domain.VendorProfile) error { return nil }
func (wVendorRepo) ListVendors(ctx context.Context, category string, limit, offset int) ([]domain.VendorProfile, error) {
	return []domain.VendorProfile{{VendorID: "v-1", BusinessName: "Katering Sedap", Category: "CATERING"}}, nil
}

type wTxRepo struct {
	domain.TransactionRepository
}

func (wTxRepo) ListTransactionsByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Transaction, error) {
	return []domain.Transaction{{ID: "tx-1", BuyerID: userID}}, nil
}
func (wTxRepo) GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error) {
	return &domain.Transaction{ID: id, BuyerID: "user-uuid-1", SellerID: "seller-1", Type: domain.TypeGoods}, nil
}
func (wTxRepo) GetGoodsDetailByTxID(ctx context.Context, txID string) (*domain.TransactionGoods, error) {
	return &domain.TransactionGoods{TransactionID: txID, ShippingCourier: "JNE", AutoConfirmDeadline: time.Now()}, nil
}

// ---- Helpers -----------------------------------------------------------------

func newWalletTestRouter(walletRepo *wWalletRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	uow := &wUoW{stores: domain.TxStores{Wallets: walletRepo, Finance: wFinRepo{}}}
	userUsecase := usecase.NewUserUsecase(uow, &hUserRepo{}, walletRepo, nil)
	withdrawalUsecase := usecase.NewWithdrawalUsecase(uow, walletRepo, nil)
	h := handlers.NewWalletHandler(userUsecase, withdrawalUsecase)

	r := gin.New()
	withUser := func(path string, hc gin.HandlerFunc) {
		r.Handle(http.MethodGet, path, func(c *gin.Context) { c.Set("user_id", "user-uuid-1"); hc(c) })
	}
	r.POST("/wallets/withdraw", func(c *gin.Context) { c.Set("user_id", "user-uuid-1"); h.RequestWithdrawalHandler(c) })
	withUser("/wallets/me", h.GetBalanceHandler)
	withUser("/wallets/me/transactions", h.GetHistoryHandler)
	withUser("/wallets/withdrawals", h.ListMyWithdrawalsHandler)
	return r
}

func doReq(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- Wallet / withdrawal handlers ----------------------------------------------

func TestWithdrawHandler_Success(t *testing.T) {
	repo := &wWalletRepo{}
	r := newWalletTestRouter(repo)

	w := doReq(r, http.MethodPost, "/wallets/withdraw",
		`{"amount":500000,"bank_name":"BCA","account_number":"1234567890","account_holder":"Budi Santoso"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	if repo.withdrawal == nil || repo.withdrawal.Fee != domain.WithdrawFeeToUser {
		t.Errorf("withdrawal request must record the gross fee %d", domain.WithdrawFeeToUser)
	}
}

func TestWithdrawHandler_InvalidPayload400(t *testing.T) {
	r := newWalletTestRouter(&wWalletRepo{})

	w := doReq(r, http.MethodPost, "/wallets/withdraw", `{"amount":0}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestWithdrawHandler_DatabaseError500(t *testing.T) {
	// A refused wallet debit (e.g. DB failure) must surface as 500, not 400.
	r := newWalletTestRouter(&wWalletRepo{debitFails: true})

	w := doReq(r, http.MethodPost, "/wallets/withdraw",
		`{"amount":500000,"bank_name":"BCA","account_number":"1234567890","account_holder":"Budi Santoso"}`)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (non-validation failures must not masquerade as 400)", w.Code)
	}
}

func TestWalletBalanceAndHistoryHandlers(t *testing.T) {
	r := newWalletTestRouter(&wWalletRepo{})

	if w := doReq(r, http.MethodGet, "/wallets/me", ""); w.Code != http.StatusOK {
		t.Fatalf("balance status = %d, want 200", w.Code)
	}
	if w := doReq(r, http.MethodGet, "/wallets/me/transactions?limit=10", ""); w.Code != http.StatusOK {
		t.Fatalf("history status = %d, want 200", w.Code)
	}
}

// ---- Catalog handlers -----------------------------------------------------------

func TestCatalogHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cu := usecase.NewCatalogUsecase(wCatRepo{}, wVendorRepo{})
	h := handlers.NewCatalogHandler(cu)
	r := gin.New()
	r.GET("/categories", h.GetCategoryCatalogHandler)
	r.GET("/vendors", h.ListMarketplaceVendorsHandler)

	if w := doReq(r, http.MethodGet, "/categories", ""); w.Code != http.StatusOK {
		t.Fatalf("categories status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if w := doReq(r, http.MethodGet, "/vendors?category=CATERING", ""); w.Code != http.StatusOK {
		t.Fatalf("vendors status = %d, want 200", w.Code)
	}
}

// ---- Transaction query handlers ---------------------------------------------------

func TestTransactionQueryHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	qu := usecase.NewTransactionQueryUsecase(wTxRepo{})
	h := handlers.NewTransactionQueryHandler(qu)
	r := gin.New()
	r.GET("/transactions", func(c *gin.Context) { c.Set("user_id", "user-uuid-1"); h.ListMyTransactionsHandler(c) })
	r.GET("/transactions/:id", func(c *gin.Context) { c.Set("user_id", "user-uuid-1"); h.GetTransactionDetailHandler(c) })
	r.GET("/transactions-stranger/:id", func(c *gin.Context) { c.Set("user_id", "stranger"); h.GetTransactionDetailHandler(c) })

	if w := doReq(r, http.MethodGet, "/transactions", ""); w.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", w.Code)
	}
	if w := doReq(r, http.MethodGet, "/transactions/tx-1", ""); w.Code != http.StatusOK {
		t.Fatalf("detail status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if w := doReq(r, http.MethodGet, "/transactions-stranger/tx-1", ""); w.Code != http.StatusForbidden {
		t.Fatalf("third-party detail status = %d, want 403", w.Code)
	}
}
