package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type DisputeHandler struct {
	usecase *usecase.DisputeUsecase
}

func NewDisputeHandler(u *usecase.DisputeUsecase) *DisputeHandler {
	return &DisputeHandler{usecase: u}
}

type OpenDisputeRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required,uuid4"`
	Reason        string  `json:"reason" binding:"required"`
	EvidenceURL   *string `json:"evidence_url,omitempty"`
}

func (h *DisputeHandler) OpenDisputeHandler(c *gin.Context) {
	var req OpenDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dispute data: " + err.Error()})
		return
	}
	raisedBy, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	dispute, err := h.usecase.OpenDispute(c.Request.Context(), req.TransactionID, raisedBy.(string), req.Reason, req.EvidenceURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"dispute": dispute})
}

func (h *DisputeHandler) AcknowledgeDisputeHandler(c *gin.Context) {
	if err := h.usecase.AcknowledgeDispute(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "dispute acknowledged", "status": domain.DisputeStatusUnderReview})
}

type ResolveDisputeRequest struct {
	Outcome domain.DisputeOutcome `json:"outcome" binding:"required"`
	Summary string                `json:"summary" binding:"required"`
}

func (h *DisputeHandler) ResolveDisputeHandler(c *gin.Context) {
	var req ResolveDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resolution data: " + err.Error()})
		return
	}
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if err := h.usecase.ResolveDispute(c.Request.Context(), c.Param("id"), adminID.(string), req.Outcome, req.Summary); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "dispute resolved", "outcome": req.Outcome})
}

func (h *DisputeHandler) GetDisputeHandler(c *gin.Context) {
	dispute, err := h.usecase.GetDispute(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dispute": dispute})
}
