package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type TransactionServicesHandler struct {
	servicesUsecase *usecase.TransactionServicesUsecase
}

func NewTransactionServicesHandler(su *usecase.TransactionServicesUsecase) *TransactionServicesHandler {
	return &TransactionServicesHandler{servicesUsecase: su}
}

type LockServicesRequest struct {
	BuyerID        string `json:"buyer_id" binding:"required,uuid4"`
	SellerID       string `json:"seller_id" binding:"required,uuid4"`
	AmountBase     int64  `json:"amount_base" binding:"required,gt=0"`
	IsRekberPay    bool   `json:"is_rekber_pay"`
	SellerTier     string `json:"seller_tier" binding:"required,oneof=GOLD SILVER BRONZE"`
	PaymentMethod  string `json:"payment_method" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *TransactionServicesHandler) LockFundsServicesHandler(c *gin.Context) {
	var req LockServicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	tx, err := h.servicesUsecase.LockFundsServices(
		c.Request.Context(), req.BuyerID, req.SellerID, req.AmountBase,
		req.IsRekberPay, req.SellerTier, req.PaymentMethod, req.IdempotencyKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Escrow jasa berhasil diinisialisasi", "data": tx})
}

func (h *TransactionServicesHandler) ReleaseMilestoneHandler(c *gin.Context) {
	var req struct {
		MilestoneID string `json:"milestone_id" binding:"required,uuid4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.servicesUsecase.ReleaseMilestoneFunds(c.Request.Context(), req.MilestoneID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Dana termin milestone jasa berhasil dicairkan ke freelancer!"})
}