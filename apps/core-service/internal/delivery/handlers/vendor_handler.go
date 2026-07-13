package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type VendorHandler struct {
	vendorUsecase *usecase.VendorUsecase
}

func NewVendorHandler(vu *usecase.VendorUsecase) *VendorHandler {
	return &VendorHandler{vendorUsecase: vu}
}

type RegisterVendorPayload struct {
	BusinessName string `json:"business_name" binding:"required"`
	Category     string `json:"category" binding:"required"` 
}

func (h *VendorHandler) RegisterVendorHandler(c *gin.Context) {
	// Ambil UUID User sah dari injector token JWT Auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak ditemukan, silakan login kembali"})
		return
	}

	var req RegisterVendorPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	// Bentuk objek entitas sesuai spesifikasi vendor.go asli Anda
	vendor := &domain.VendorProfile{
		VendorID:     userID.(string), // VendorID terikat murni ke UserID pengaju
		BusinessName: req.BusinessName,
		Category:     req.Category,
		IsVerified:   false,
	}

	if err := h.vendorUsecase.RegisterVendorProfile(c.Request.Context(), vendor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Pengajuan profil bisnis vendor berhasil direkam, menunggu kurasi Admin",
	})
}