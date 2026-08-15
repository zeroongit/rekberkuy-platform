package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/usecase"
)

type TransactionQueryHandler struct {
	queryUsecase *usecase.TransactionQueryUsecase
}

func NewTransactionQueryHandler(qu *usecase.TransactionQueryUsecase) *TransactionQueryHandler {
	return &TransactionQueryHandler{queryUsecase: qu}
}

func (h *TransactionQueryHandler) ListMyTransactionsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session or user not logged in"})
		return
	}

	limit, offset := pagination(c)
	txs, err := h.queryUsecase.ListUserTransactions(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": txs})
}

func (h *TransactionQueryHandler) GetTransactionDetailHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session or user not logged in"})
		return
	}

	txID := c.Param("id")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction id is required"})
		return
	}

	detail, err := h.queryUsecase.GetTransactionDetail(c.Request.Context(), userID.(string), txID)
	if err != nil {
		if errors.Is(err, usecase.ErrNotTransactionParty) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": detail})
}

// pagination reads ?limit & ?offset with sane defaults/caps.
func pagination(c *gin.Context) (limit, offset int) {
	limit = 20
	offset = 0
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 && v <= 100 {
		limit = v
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v > 0 {
		offset = v
	}
	return limit, offset
}
