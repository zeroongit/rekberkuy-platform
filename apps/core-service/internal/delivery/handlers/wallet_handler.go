package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"rekberkuy/core-service/internal/usecase"
)

type WalletHandler struct {
	userUsecase *usecase.UserUsecase // Targets the usecase, following Clean Architecture rules
}

func NewWalletHandler(uu *usecase.UserUsecase) *WalletHandler {
	return &WalletHandler{userUsecase: uu}
}

type TopUpRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

// CreateTopUpHandler creates a top-up draft + Midtrans Snap transaction, then
// returns the snap token & redirect URL so the user can complete payment.
func (h *WalletHandler) CreateTopUpHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid top-up amount: " + err.Error()})
		return
	}

	txLog, snap, err := h.userUsecase.TopUpWallet(c.Request.Context(), userID.(string), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":       "success",
		"transaction":  txLog,
		"snap_token":   snap.Token,
		"redirect_url": snap.RedirectURL,
	})
}
