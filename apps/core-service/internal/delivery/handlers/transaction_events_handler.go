package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type TransactionEventsHandler struct {
	eventsUsecase *usecase.TransactionEventsUsecase
}

func NewTransactionEventsHandler(eu *usecase.TransactionEventsUsecase) *TransactionEventsHandler {
	return &TransactionEventsHandler{eventsUsecase: eu}
}

type VendorAllocationPayload struct {
	VendorID        string `json:"vendor_id" binding:"required,uuid4"`
	AllocatedAmount int64  `json:"allocated_amount" binding:"required,gt=0"`
}

type EventsDetailPayload struct {
	SubSubCategoryID    uint64                     `json:"sub_sub_category_id" binding:"required,gt=0"`
	EventName           string                     `json:"event_name" binding:"required"`
	EventStartTime      string                     `json:"event_start_time" binding:"required"` // RFC3339
	EventEndTime        string                     `json:"event_end_time" binding:"required"`   // RFC3339
	TicketQuantityTotal int                        `json:"ticket_quantity_total" binding:"min=0"`
	VendorAllocations   []VendorAllocationPayload  `json:"vendor_allocations" binding:"dive"`
}

type LockEventsRequest struct {
	BuyerID        string                `json:"buyer_id" binding:"required,uuid4"`
	SellerID       string                `json:"seller_id" binding:"required,uuid4"`
	AmountBase     int64                 `json:"amount_base" binding:"required,gt=0"`
	IsRekberPay    bool                  `json:"is_rekber_pay"`
	SellerTier     string                `json:"seller_tier" binding:"required,oneof=GOLD SILVER BRONZE"`
	PaymentMethod  string                `json:"payment_method" binding:"required"`
	IdempotencyKey string                `json:"idempotency_key" binding:"required"`
	Details        *EventsDetailPayload  `json:"details"`
}

func (h *TransactionEventsHandler) LockFundsEventsHandler(c *gin.Context) {
	var req LockEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	var detail *domain.TransactionEvents
	var allocations []domain.EventVendorAllocation
	if req.Details != nil {
		startTime, err := time.Parse(time.RFC3339, req.Details.EventStartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "event_start_time must be an RFC3339 timestamp"})
			return
		}
		endTime, err := time.Parse(time.RFC3339, req.Details.EventEndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "event_end_time must be an RFC3339 timestamp"})
			return
		}
		detail = &domain.TransactionEvents{
			SubSubCategoryID:    req.Details.SubSubCategoryID,
			EventName:           req.Details.EventName,
			EventStartTime:      startTime,
			EventEndTime:        endTime,
			TicketQuantityTotal: req.Details.TicketQuantityTotal,
		}
		for _, a := range req.Details.VendorAllocations {
			allocations = append(allocations, domain.EventVendorAllocation{
				VendorID:        a.VendorID,
				AllocatedAmount: a.AllocatedAmount,
				Status:          domain.VendorAllocationPledged,
			})
		}
	}

	tx, err := h.eventsUsecase.LockFundsEvents(
		c.Request.Context(), req.BuyerID, req.SellerID, req.AmountBase,
		req.IsRekberPay, req.SellerTier, req.PaymentMethod, req.IdempotencyKey, detail, allocations,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Event escrow successfully initialized", "data": tx})
}

// SubmitVendorInvoicePayload is a vendor's payment-claim invoice for an event.
// Per the platform's IPFS proof-storage rule, invoice_file_url is expected to
// be the gateway URL of the invoice uploaded to IPFS.
type SubmitVendorInvoicePayload struct {
	VendorUserID        *string `json:"vendor_user_id" binding:"omitempty,uuid4"` // nil = external vendor
	VendorName          string  `json:"vendor_name" binding:"required"`
	VendorBankName      string  `json:"vendor_bank_name" binding:"required"`
	VendorAccountNumber string  `json:"vendor_account_number" binding:"required"`
	AmountRequested     int64   `json:"amount_requested" binding:"required,gt=0"`
	ExpenseDescription  string  `json:"expense_description" binding:"required"`
	InvoiceFileURL      string  `json:"invoice_file_url" binding:"required,url"`
	PayoutPhase         string  `json:"payout_phase"` // default FINAL_SETTLEMENT
}

func (h *TransactionEventsHandler) SubmitVendorInvoiceHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session or user not logged in"})
		return
	}

	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event transaction id is required"})
		return
	}

	var req SubmitVendorInvoicePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice payload: " + err.Error()})
		return
	}

	payout := &domain.EventVendorPayout{
		TransactionID:       transactionID,
		VendorUserID:        req.VendorUserID,
		VendorName:          req.VendorName,
		VendorBankName:      req.VendorBankName,
		VendorAccountNumber: req.VendorAccountNumber,
		AmountRequested:     req.AmountRequested,
		ExpenseDescription:  req.ExpenseDescription,
		InvoiceFileURL:      req.InvoiceFileURL,
		PayoutPhase:         req.PayoutPhase,
	}

	if err := h.eventsUsecase.SubmitEventVendorInvoice(c.Request.Context(), userID.(string), payout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Vendor invoice submitted and awaits admin/EO payout processing",
		"data":    payout,
	})
}

func (h *TransactionEventsHandler) ProcessEventVendorPayoutHandler(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transaction_id" binding:"required,uuid4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventsUsecase.ProcessEventVendorPayouts(c.Request.Context(), req.TransactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Event escrow funds successfully split to all field vendors!"})
}

func (h *TransactionEventsHandler) ReleaseEventMilestoneHandler(c *gin.Context) {
	var req struct {
		PayoutID string `json:"payout_id" binding:"required,uuid4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventsUsecase.ReleaseEventMilestonePayout(c.Request.Context(), req.PayoutID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invoice-based event milestone funds successfully disbursed!"})
}
