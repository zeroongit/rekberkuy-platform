package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	userUsecase *usecase.UserUsecase // Menggunakan usecase user yang memegang kendali transaksi dompet
}

func NewWalletHandler(uu *usecase.UserUsecase) *WalletHandler {
	return &WalletHandler{userUsecase: uu}
}

// CreateTopUpHandler memicu pembuatan invoice pembayaran/Snap Token Midtrans
func (h *WalletHandler) CreateTopUpHandler(c *gin.Context) {
	userID, _ := c.Get("user_id") // Diambil otomatis dari JWT Auth Middleware kita

	c.JSON(http.StatusCreated, gin.H{
		"status":        "success",
		"user_id":       userID,
		"snap_token":    "midtrans-snap-token-simulation-xyz123",
		"redirect_url":  "https://app.sandbox.midtrans.com/snap/v2/vtweb/midtrans-snap-token-simulation-xyz123",
		"message":       "Sistem BFF Tersembunyi: Token pembayaran berhasil diterbitkan",
	})
}