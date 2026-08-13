package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
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
	// Get the legitimate User UUID from the JWT Auth middleware token injector
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found, please log in again"})
		return
	}

	var req RegisterVendorPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	// Build the entity object per your actual vendor.go specification
	vendor := &domain.VendorProfile{
		VendorID:     userID.(string), // VendorID is purely bound to the requester's UserID
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
		"message": "Vendor business profile submission successfully recorded, awaiting Admin curation",
	})
}
