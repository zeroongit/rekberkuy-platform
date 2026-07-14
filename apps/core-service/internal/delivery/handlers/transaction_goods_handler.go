package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type TransactionGoodsHandler struct {
	goodsUsecase *usecase.TransactionGoodsUsecase
}

func NewTransactionGoodsHandler(gu *usecase.TransactionGoodsUsecase) *TransactionGoodsHandler {
	return &TransactionGoodsHandler{goodsUsecase: gu}
}

type LockGoodsRequest struct {
	BuyerID        string `json:"buyer_id" binding:"required,uuid4"`
	SellerID       string `json:"seller_id" binding:"required,uuid4"`
	AmountBase     int64  `json:"amount_base" binding:"required,gt=0"`
	IsRekberPay    bool   `json:"is_rekber_pay"`
	SellerTier     string `json:"seller_tier" binding:"required,oneof=GOLD SILVER BRONZE"`
	ShippingFee    int64  `json:"shipping_fee" binding:"min=0"`
	PaymentMethod  string `json:"payment_method" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *TransactionGoodsHandler) LockFundsGoodsHandler(c *gin.Context) {
	var req LockGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	tx, err := h.goodsUsecase.LockFundsGoods(
		c.Request.Context(), req.BuyerID, req.SellerID, req.AmountBase,
		req.IsRekberPay, req.SellerTier, req.ShippingFee, req.PaymentMethod, req.IdempotencyKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Escrow barang diinisialisasi", "data": tx})
}

func (h *TransactionGoodsHandler) ReleaseGoodsHandler(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transaction_id" binding:"required,uuid4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.goodsUsecase.ReleaseFundsGoods(c.Request.Context(), req.TransactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Dana escrow barang dikreditkan ke dompet penjual"})
}