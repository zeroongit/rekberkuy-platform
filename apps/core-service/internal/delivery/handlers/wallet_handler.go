package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	userUsecase *usecase.UserUsecase // Menembak usecase, mematuhi aturan Clean Architecture
}

func NewWalletHandler(uu *usecase.UserUsecase) *WalletHandler {
	return &WalletHandler{userUsecase: uu}
}

type TopUpRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

// CreateTopUpHandler mengamankan aliran request finansial
func (h *WalletHandler) CreateTopUpHandler(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid atau kedaluwarsa"})
		return
	}

	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nominal top-up tidak valid: " + err.Error()})
		return
	}

	// 2. Oper data murni ke layer usecase untuk divalidasi secara terisolasi
	txLog, err := h.userUsecase.TopUpWallet(c.Request.Context(), userID.(string), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


	c.JSON(http.StatusCreated, gin.H{
		"status":       "success",
		"transaction":  txLog,
		"snap_token":   "midtrans-snap-token-real-mvp-xyz",
		"redirect_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/midtrans-snap-token-real-mvp-xyz",
	})
}