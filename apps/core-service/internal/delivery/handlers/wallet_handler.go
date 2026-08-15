package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"rekberkuy/core-service/internal/usecase"
)

type WalletHandler struct {
	userUsecase      *usecase.UserUsecase // Targets the usecase, following Clean Architecture rules
	withdrawalUsecase *usecase.WithdrawalUsecase
}

func NewWalletHandler(uu *usecase.UserUsecase, wu *usecase.WithdrawalUsecase) *WalletHandler {
	return &WalletHandler{userUsecase: uu, withdrawalUsecase: wu}
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

// GetBalanceHandler returns the caller's RekberPay balance + frozen flag.
func (h *WalletHandler) GetBalanceHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}
	wallet, err := h.userUsecase.GetWalletBalance(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": wallet})
}

// GetHistoryHandler returns the caller's wallet ledger (newest first).
func (h *WalletHandler) GetHistoryHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}
	limit, offset := pagination(c)
	history, err := h.userUsecase.GetWalletHistory(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": history})
}

type WithdrawRequest struct {
	Amount         int64  `json:"amount" binding:"required,gt=0"`
	BankName       string `json:"bank_name" binding:"required"`
	AccountNumber  string `json:"account_number" binding:"required"`
	AccountHolder  string `json:"account_holder" binding:"required"`
}

// RequestWithdrawalHandler debits the wallet (amount + flat fee) and records a
// PENDING withdrawal request. The bank transfer completes out-of-band.
func (h *WalletHandler) RequestWithdrawalHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid withdrawal payload: " + err.Error()})
		return
	}

	request, err := h.withdrawalUsecase.RequestWithdrawal(
		c.Request.Context(), userID.(string), req.Amount, req.BankName, req.AccountNumber, req.AccountHolder,
	)
	if err != nil {
		// Validation problems are the caller's fault (400); anything else
		// (wallet debit refused by the DB, UoW failure) is a server error.
		if errors.Is(err, usecase.ErrInvalidWithdrawalReqt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Withdrawal request recorded", "data": request})
}

// ListMyWithdrawalsHandler returns the caller's own withdrawal requests.
func (h *WalletHandler) ListMyWithdrawalsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}
	limit, offset := pagination(c)
	withdrawals, err := h.withdrawalUsecase.ListMyWithdrawals(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": withdrawals})
}

// ListPendingWithdrawalsHandler serves the admin disbursement queue.
func (h *WalletHandler) ListPendingWithdrawalsHandler(c *gin.Context) {
	limit, offset := pagination(c)
	withdrawals, err := h.withdrawalUsecase.ListPendingWithdrawals(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": withdrawals})
}

// MarkWithdrawalDisbursedHandler lets an admin confirm the bank transfer done.
// Optional body {"midtrans_fee": <int>} supplies the REAL disbursement cost —
// the ledger true-ups the difference against the booked estimate automatically.
func (h *WalletHandler) MarkWithdrawalDisbursedHandler(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	withdrawalID := c.Param("id")
	if withdrawalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal id is required"})
		return
	}

	var req struct {
		MidtransFee *int64 `json:"midtrans_fee" binding:"omitempty,gte=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid disbursement payload: " + err.Error()})
		return
	}

	if err := h.withdrawalUsecase.MarkWithdrawalDisbursed(c.Request.Context(), withdrawalID, adminID.(string), req.MidtransFee); err != nil {
		switch {
		case errors.Is(err, usecase.ErrWithdrawalAlreadyPaid):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrInvalidWithdrawalReqt):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Withdrawal marked as paid"})
}
