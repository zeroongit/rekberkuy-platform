package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type TransactionServicesHandler struct {
	servicesUsecase *usecase.TransactionServicesUsecase
}

func NewTransactionServicesHandler(su *usecase.TransactionServicesUsecase) *TransactionServicesHandler {
	return &TransactionServicesHandler{servicesUsecase: su}
}

type MilestonePayload struct {
	Title  string `json:"title" binding:"required"`
	Amount int64  `json:"amount" binding:"required,gt=0"`
}

type ServicesDetailPayload struct {
	SubSubCategoryID uint64              `json:"sub_sub_category_id" binding:"required,gt=0"`
	ProjectDeadline  string              `json:"project_deadline" binding:"required"` // RFC3339
	BriefDescription string              `json:"brief_description" binding:"required"`
	Milestones       []MilestonePayload  `json:"milestones" binding:"required,gt=0,dive"`
}

type LockServicesRequest struct {
	BuyerID        string                  `json:"buyer_id" binding:"required,uuid4"`
	SellerID       string                  `json:"seller_id" binding:"required,uuid4"`
	AmountBase     int64                   `json:"amount_base" binding:"required,gt=0"`
	IsRekberPay    bool                    `json:"is_rekber_pay"`
	SellerTier     string                  `json:"seller_tier" binding:"required,oneof=GOLD SILVER BRONZE"`
	PaymentMethod  string                  `json:"payment_method" binding:"required"`
	IdempotencyKey string                  `json:"idempotency_key" binding:"required"`
	Details        *ServicesDetailPayload  `json:"details"`
}

func (h *TransactionServicesHandler) LockFundsServicesHandler(c *gin.Context) {
	var req LockServicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	var detail *domain.TransactionServices
	var milestones []domain.ServiceMilestone
	if req.Details != nil {
		deadline, err := time.Parse(time.RFC3339, req.Details.ProjectDeadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_deadline must be an RFC3339 timestamp"})
			return
		}
		detail = &domain.TransactionServices{
			SubSubCategoryID: req.Details.SubSubCategoryID,
			ProjectDeadline:  deadline,
			BriefDescription: req.Details.BriefDescription,
		}
		for _, m := range req.Details.Milestones {
			milestones = append(milestones, domain.ServiceMilestone{Title: m.Title, Amount: m.Amount})
		}
	}

	tx, err := h.servicesUsecase.LockFundsServices(
		c.Request.Context(), req.BuyerID, req.SellerID, req.AmountBase,
		req.IsRekberPay, req.SellerTier, req.PaymentMethod, req.IdempotencyKey, detail, milestones,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Service escrow successfully initialized", "data": tx})
}

func (h *TransactionServicesHandler) ReleaseMilestoneHandler(c *gin.Context) {
	var req struct {
		MilestoneID string `json:"milestone_id" binding:"required,uuid4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.servicesUsecase.ReleaseMilestoneFunds(c.Request.Context(), req.MilestoneID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Service milestone funds successfully disbursed to freelancer!"})
}
