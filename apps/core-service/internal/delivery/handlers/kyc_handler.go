package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type KYCHandler struct {
	kycUsecase *usecase.KYCUsecase
}

func NewKYCHandler(ku *usecase.KYCUsecase) *KYCHandler {
	return &KYCHandler{kycUsecase: ku}
}

type KYCRequestPayload struct {
	TargetRole   string `json:"target_role" binding:"required,oneof=VERIFIED_MERCHANT VERIFIED_VENDOR EVENT_ORGANIZER"`
	IDCardNumber string `json:"id_card_number" binding:"required,numeric,len=16"`
	IDCardURL    string `json:"id_card_url" binding:"required,url"`
	SelfieURL    string `json:"selfie_url" binding:"required,url"`
}

func (h *KYCHandler) SubmitKYCHandler(c *gin.Context) {
	userID, exists := c.Get("user_id") // Protected by JWT Auth middleware
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session or user not logged in"})
		return
	}

	var req KYCRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid identity payload: " + err.Error()})
		return
	}

	targetUserRole := domain.UserRole(req.TargetRole)
	err := h.kycUsecase.SubmitUserKYC(
		c.Request.Context(),
		userID.(string),
		targetUserRole,
		req.IDCardNumber,
		req.IDCardURL,
		req.SelfieURL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "KYC documents successfully uploaded, your verification queue is being processed by admin",
	})
}
