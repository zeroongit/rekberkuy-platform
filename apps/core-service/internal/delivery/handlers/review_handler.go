package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/usecase"
)

type ReviewHandler struct {
	usecase *usecase.ReviewUsecase
}

func NewReviewHandler(u *usecase.ReviewUsecase) *ReviewHandler {
	return &ReviewHandler{usecase: u}
}

type CreateReviewRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required,uuid4"`
	Rating        int     `json:"rating" binding:"required,min=1,max=5"`
	Comment       *string `json:"comment,omitempty"`
}

func (h *ReviewHandler) CreateReviewHandler(c *gin.Context) {
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review data: " + err.Error()})
		return
	}
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	review, err := h.usecase.CreateReview(c.Request.Context(), reviewerID.(string), req.TransactionID, req.Rating, req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"review": review})
}

func (h *ReviewHandler) GetReviewsForUserHandler(c *gin.Context) {
	userID := c.Param("id")
	reviews, err := h.usecase.GetReviewsForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}
