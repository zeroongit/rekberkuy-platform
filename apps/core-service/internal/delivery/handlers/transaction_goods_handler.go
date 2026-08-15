package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type TransactionGoodsHandler struct {
	goodsUsecase *usecase.TransactionGoodsUsecase
}

func NewTransactionGoodsHandler(gu *usecase.TransactionGoodsUsecase) *TransactionGoodsHandler {
	return &TransactionGoodsHandler{goodsUsecase: gu}
}

type GoodsDetailPayload struct {
	SubSubCategoryID    uint64 `json:"sub_sub_category_id" binding:"required,gt=0"`
	ShippingCourier     string `json:"shipping_courier" binding:"required"`
	ShippingAddress     string `json:"shipping_address" binding:"required"`
	AutoConfirmDeadline string `json:"auto_confirm_deadline" binding:"required"` // RFC3339
}

type LockGoodsRequest struct {
	BuyerID        string             `json:"buyer_id" binding:"required,uuid4"`
	SellerID       string             `json:"seller_id" binding:"required,uuid4"`
	AmountBase     int64              `json:"amount_base" binding:"required,gt=0"`
	IsRekberPay    bool               `json:"is_rekber_pay"`
	SellerTier     string             `json:"seller_tier" binding:"required,oneof=GOLD SILVER BRONZE"`
	ShippingFee    int64              `json:"shipping_fee" binding:"min=0"`
	PaymentMethod  string             `json:"payment_method" binding:"required"`
	IdempotencyKey string             `json:"idempotency_key" binding:"required"`
	Details        *GoodsDetailPayload `json:"details"`
}

func (h *TransactionGoodsHandler) LockFundsGoodsHandler(c *gin.Context) {
	var req LockGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	var detail *domain.TransactionGoods
	if req.Details != nil {
		deadline, err := time.Parse(time.RFC3339, req.Details.AutoConfirmDeadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "auto_confirm_deadline must be an RFC3339 timestamp"})
			return
		}
		detail = &domain.TransactionGoods{
			SubSubCategoryID: req.Details.SubSubCategoryID,
			ShippingCourier:  req.Details.ShippingCourier,
			ShippingAddress:  req.Details.ShippingAddress,
			AutoConfirmDeadline: deadline,
		}
	}

	tx, err := h.goodsUsecase.LockFundsGoods(
		c.Request.Context(), req.BuyerID, req.SellerID, req.AmountBase,
		req.IsRekberPay, req.SellerTier, req.ShippingFee, req.PaymentMethod, req.IdempotencyKey, detail,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Goods escrow initialized", "data": tx})
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
	c.JSON(http.StatusOK, gin.H{"message": "Goods escrow funds credited to seller wallet"})
}
