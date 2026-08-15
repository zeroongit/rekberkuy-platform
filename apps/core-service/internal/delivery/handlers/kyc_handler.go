package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

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

// KYCReviewPayload is the admin's decision body for reviewing a submission.
// The AI score stored on the submission is a REFERENCE only — the approve
// decision recorded here is the final, human-owned verdict.
type KYCReviewPayload struct {
	Approve bool   `json:"approve"`
	Notes   string `json:"notes"`
}

func (h *KYCHandler) ReviewKYCHandler(c *gin.Context) {
	adminID, exists := c.Get("user_id") // Protected by JWT Auth middleware (RoleAdmin)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session or admin not logged in"})
		return
	}

	kycID := c.Param("id")
	if kycID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KYC submission id is required"})
		return
	}

	var req KYCReviewPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review payload: " + err.Error()})
		return
	}

	decision := domain.KYCRejected
	if req.Approve {
		decision = domain.KYCApproved
	}

	reviewed, err := h.kycUsecase.ReviewKYC(c.Request.Context(), adminID.(string), kycID, decision, req.Notes)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrKYCAlreadyReviewed):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrKYCInvalidDecision):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "KYC review decision recorded",
		"submission":  reviewed,
	})
}

func (h *KYCHandler) GetPendingKYCsHandler(c *gin.Context) {
	pending, err := h.kycUsecase.ListPendingKYCs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": pending})
}

func (h *KYCHandler) GetKYCDetailHandler(c *gin.Context) {
	kycID := c.Param("id")
	if kycID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KYC submission id is required"})
		return
	}

	submission, err := h.kycUsecase.GetKYCByID(c.Request.Context(), kycID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": submission})
}
