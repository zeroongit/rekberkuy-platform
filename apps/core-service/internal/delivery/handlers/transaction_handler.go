package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	txUsecase *usecase.TransactionUsecase 
}

func NewTransactionHandler(tu *usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{txUsecase: tu}
}

// ============================================================================
// 📨 SEKTOR DTO (DATA TRANSFER OBJECT) REQUEST PAYLOAD
// ============================================================================

type LockFundsRequest struct {
	BuyerID        string `json:"buyer_id" binding:"required,uuid4"`
	SellerID       string `json:"seller_id" binding:"required,uuid4"`
	AmountBase     int64  `json:"amount_base" binding:"required,gt=0"`
	RekberType     string `json:"rekber_type" binding:"required,oneof=GOODS SERVICES EVENTS"`
	IsRekberPay    bool   `json:"is_rekber_pay"`
	SellerTier     string `json:"seller_tier" binding:"required,oneof=GOLD SILVER BRONZE"`
	ShippingFee    int64  `json:"shipping_fee" binding:"min=0"`
	PaymentMethod  string `json:"payment_method" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

type WebhookMidtransRequest struct {
	TransactionID string `json:"transaction_id" binding:"required,uuid4"`
}

type ReleaseFundsRequest struct {
	TransactionID string `json:"transaction_id" binding:"required,uuid4"`
}

// ============================================================================
// 📡 SEKTOR HTTP HANDLER ENDPOINTS
// ============================================================================

// LockFundsAwalHandler menangani POST /api/v1/transactions (FUNDS_LOCKED / WAITING_PAYMENT)
func (h *TransactionHandler) LockFundsAwalHandler(c *gin.Context) {
	var req LockFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	// Konversi string request ke tipe domain RekberType yang valid
	rekberType := domain.RekberType(req.RekberType)

	tx, err := h.txUsecase.LockFundsAwal(
		c.Request.Context(),
		req.BuyerID,
		req.SellerID,
		req.AmountBase,
		rekberType,
		req.IsRekberPay,
		req.SellerTier,
		req.ShippingFee,
		req.PaymentMethod,
		req.IdempotencyKey,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaksi escrow berhasil diinisialisasi, menunggu pembayaran",
		"data":    tx,
	})
}

// MidtransWebhookHandler menangani POST /api/v1/webhooks/midtrans
func (h *TransactionHandler) MidtransWebhookHandler(c *gin.Context) {
	var req WebhookMidtransRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	err := h.txUsecase.ConfirmPaymentWebhookMidtrans(c.Request.Context(), req.TransactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook berhasil diproses, dana berhasil dikunci di escrow RekberKuy",
	})
}

// ReleaseFundsHandler menangani POST /api/v1/transactions/release
func (h *TransactionHandler) ReleaseFundsHandler(c *gin.Context) {
	var req ReleaseFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	err := h.txUsecase.ReleaseFundsSelesai(c.Request.Context(), req.TransactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dana escrow berhasil dilepas dan dikreditkan ke dompet penjual",
	})
}


type ReleaseMilestoneRequest struct {
	MilestoneID string `json:"milestone_id" binding:"required,uuid4"`
}


func (h *TransactionHandler) ReleaseMilestoneHandler(c *gin.Context) {
	var req ReleaseMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	err := h.txUsecase.ReleaseMilestoneFunds(c.Request.Context(), req.MilestoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dana jatah milestone termin berhasil dicairkan ke dompet freelancer!",
	})
}