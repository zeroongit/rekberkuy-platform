package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// MidtransWebhookHandler handles Midtrans (payment gateway) webhook notifications.
// It verifies the SignatureKey then routes to the appropriate usecase based on the
// order_id prefix: top-up (REKBERKUY-TOPUP-), goods (REKBERKUY-GOODS-),
// services (REKBERKUY-SERVICE-), or event (REKBERKUY-EVENT-).
type MidtransWebhookHandler struct {
	midtrans        domain.MidtransClient
	userUsecase     *usecase.UserUsecase
	goodsUsecase    *usecase.TransactionGoodsUsecase
	servicesUsecase *usecase.TransactionServicesUsecase
	eventsUsecase   *usecase.TransactionEventsUsecase
}

func NewMidtransWebhookHandler(midtrans domain.MidtransClient, uu *usecase.UserUsecase, gu *usecase.TransactionGoodsUsecase, su *usecase.TransactionServicesUsecase, eu *usecase.TransactionEventsUsecase) *MidtransWebhookHandler {
	return &MidtransWebhookHandler{midtrans: midtrans, userUsecase: uu, goodsUsecase: gu, servicesUsecase: su, eventsUsecase: eu}
}

// isSuccessStatus determines whether the Midtrans notification status means payment success.
func isSuccessStatus(transactionStatus, fraudStatus string) bool {
	if transactionStatus != "capture" && transactionStatus != "settlement" {
		return false
	}
	// fraudStatus "deny"/"challenge" is considered not yet successful; empty/"accept" = success.
	return fraudStatus == "" || fraudStatus == "accept"
}

// NotificationHandler is the webhook endpoint called by Midtrans.
func (h *MidtransWebhookHandler) NotificationHandler(c *gin.Context) {
	if h.midtrans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment gateway not configured"})
		return
	}

	var notif domain.MidtransNotification
	if err := c.ShouldBindJSON(&notif); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification payload: " + err.Error()})
		return
	}

	// 1. SignatureKey verification is mandatory (anti webhook forgery).
	if !h.midtrans.VerifySignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, notif.SignatureKey) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}

	success := isSuccessStatus(notif.TransactionStatus, notif.FraudStatus)

	// 2. Route based on the order_id prefix.
	switch {
	case strings.HasPrefix(notif.OrderID, domain.OrderPrefixTopUp):
		if !success {
			_ = h.userUsecase.MarkTopUpFailed(c.Request.Context(), notif.OrderID)
			c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "top-up not yet successful"})
			return
		}
		if err := h.userUsecase.ConfirmTopUp(c.Request.Context(), notif.OrderID, notif.TransactionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "top-up confirmed"})

	case strings.HasPrefix(notif.OrderID, domain.OrderPrefixGoods):
		if !success {
			c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "goods payment not yet successful"})
			return
		}
		if err := h.goodsUsecase.ConfirmPaymentGoods(c.Request.Context(), notif.OrderID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "goods payment confirmed"})

	case strings.HasPrefix(notif.OrderID, domain.OrderPrefixService):
		if !success {
			c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "services payment not yet successful"})
			return
		}
		if err := h.servicesUsecase.ConfirmPaymentServices(c.Request.Context(), notif.OrderID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "services payment confirmed"})

	case strings.HasPrefix(notif.OrderID, domain.OrderPrefixEvent):
		if !success {
			c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "event payment not yet successful"})
			return
		}
		if err := h.eventsUsecase.ConfirmPaymentEvents(c.Request.Context(), notif.OrderID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "event payment confirmed"})

	default:
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "order type not handled"})
	}
}
