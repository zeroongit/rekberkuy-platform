package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/usecase"
)

type DisbursementHandler struct {
	usecase *usecase.DisbursementUsecase
}

func NewDisbursementHandler(u *usecase.DisbursementUsecase) *DisbursementHandler {
	return &DisbursementHandler{usecase: u}
}

// MarkDisbursedHandler records that an Admin completed an external vendor's bank
// payout out-of-band. Route is ADMIN-only; the :id path param is the payout id.
func (h *DisbursementHandler) MarkDisbursedHandler(c *gin.Context) {
	payoutID := c.Param("id")
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if err := h.usecase.MarkExternalPayoutDisbursed(c.Request.Context(), payoutID, adminID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "vendor payout marked disbursed",
		"payout_id": payoutID,
	})
}
