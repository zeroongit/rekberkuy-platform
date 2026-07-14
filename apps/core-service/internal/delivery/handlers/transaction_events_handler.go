package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type TransactionEventsHandler struct {
	eventsUsecase *usecase.TransactionEventsUsecase
}

func NewTransactionEventsHandler(eu *usecase.TransactionEventsUsecase) *TransactionEventsHandler {
	return &TransactionEventsHandler{eventsUsecase: eu}
}

type LockEventsRequest struct {
	BuyerID        string `json:"buyer_id" binding:"required,uuid4"`
	SellerID       string `json:"seller_id" binding:"required,uuid4"`
	AmountBase     int64  `json:"amount_base" binding:"required,gt=0"`
	IsRekberPay    bool   `json:"is_rekber_pay"`
	SellerTier     string `json:"seller_tier" binding:"required,oneof=GOLD SILVER BRONZE"`
	PaymentMethod  string `json:"payment_method" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *TransactionEventsHandler) LockFundsEventsHandler(c *gin.Context) {
	var req LockEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	tx, err := h.eventsUsecase.LockFundsEvents(
		c.Request.Context(), req.BuyerID, req.SellerID, req.AmountBase,
		req.IsRekberPay, req.SellerTier, req.PaymentMethod, req.IdempotencyKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Escrow event berhasil diinisialisasi", "data": tx})
}

func (h *TransactionEventsHandler) ProcessEventVendorPayoutHandler(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transaction_id" binding:"required,uuid4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventsUsecase.ProcessEventVendorPayouts(c.Request.Context(), req.TransactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Dana escrow event berhasil dipecah ke seluruh vendor lapangan!"})
}

func (h *TransactionEventsHandler) ReleaseEventMilestoneHandler(c *gin.Context) {
	var req struct {
		PayoutID string `json:"payout_id" binding:"required,uuid4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventsUsecase.ReleaseEventMilestonePayout(c.Request.Context(), req.PayoutID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Dana termin event berbasis invoice berhasil dicairkan!"})
}