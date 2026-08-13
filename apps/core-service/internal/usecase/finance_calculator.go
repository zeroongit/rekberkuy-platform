package usecase

import (
	"rekberkuy/core-service/internal/domain"
)

type FinanceCalculator struct{}

func NewFinanceCalculator() *FinanceCalculator {
	return &FinanceCalculator{}
}

// ============================================================================
// 📦 SECTOR 1: TRANSACTION FEE CALCULATION (UPFRONT & RELEASE)
// ============================================================================

// CalculateBuyerServiceFee calculates the buyer protection fee at the start of a transaction
func (c *FinanceCalculator) CalculateBuyerServiceFee(rekberType domain.RekberType, amountBase int64, isRekberPay bool, sellerTier string) int64 {
	if rekberType == domain.TypeServices {
		return 0
	}

	if rekberType == domain.TypeGoods {
		switch sellerTier {
		case "GOLD":
			return amountBase * 8 / 100
		case "SILVER":
			return amountBase * 4 / 100
		default: // BRONZE
			if isRekberPay {
				return domain.FeeGoodsRekberPay
			}
			return domain.FeeGoodsNonRekberPay
		}
	}
	return 0
}

// CalculateSellerServiceFee calculates the merchant commission deduction when funds are released
func (c *FinanceCalculator) CalculateSellerServiceFee(rekberType domain.RekberType, amountToRelease int64, sellerTier string) int64 {
	if rekberType == domain.TypeServices {
		// Services commission is a flat 5%, independent of the seller's CRM tier.
		return amountToRelease * 5 / 100
	}

	if rekberType == domain.TypeGoods {
		switch sellerTier {
		case "GOLD":
			return amountToRelease * 3 / 100
		case "SILVER":
			return amountToRelease * 6 / 100
		default: // BRONZE
			return amountToRelease * 10 / 100
		}
	}
	return 0
}

// ============================================================================
// 🎪 SECTOR 2: EVENT LINE FINANCIAL AUDIT (POST-EVENT AUDIT ENGINE)
// ============================================================================

// CalculateEventAudit calculates the surplus distribution of Event Organizer funds
func (c *FinanceCalculator) CalculateEventAudit(totalEscrowLocked int64, vendorsBill []domain.EventVendorAllocation, eoTier string) domain.EventAuditResult {
	platformFee := totalEscrowLocked * 5 / 100
	maxVendorPool := totalEscrowLocked - platformFee

	var totalActualSpent int64 = 0
	for _, bill := range vendorsBill {
		totalActualSpent += bill.ActualPaidAmount
	}

	if totalActualSpent > maxVendorPool {
		totalActualSpent = maxVendorPool
	}

	netSurplus := maxVendorPool - totalActualSpent
	if netSurplus < 0 {
		netSurplus = 0
	}

	var bonusToEO int64 = 0
	var refundToPeserta int64 = 0

	// A surplus of at most Rp 500,000 is paid 100% to the EO as a performance
	// bonus, regardless of the event scale. Larger surpluses are split between
	// the EO bonus (tier-based percentage) and an auto-refund to participants.
	if netSurplus <= 500000 {
		bonusToEO = netSurplus
		refundToPeserta = 0
	} else {
		// Call a separate function to get the bonus percentage based on tier
		eoBonusPercent := c.getEventBonusPercentage(totalEscrowLocked, eoTier)
		bonusToEO = netSurplus * int64(eoBonusPercent) / 100
		refundToPeserta = netSurplus - bonusToEO
	}

	if refundToPeserta < 0 {
		refundToPeserta = 0
	}

	return domain.EventAuditResult{
		PlatformFee:     platformFee,
		AmountToVendor:  totalActualSpent,
		BonusToEO:       bonusToEO,
		RefundToPeserta: refundToPeserta,
	}
}

// Internal helper to separate the event bonus percentage logic.
// Returns the EO's share of a macro surplus as a whole-number percent
// (e.g. 15 for 15%). Event scale decides the band; tier decides the rate.
func (c *FinanceCalculator) getEventBonusPercentage(totalEscrow int64, eoTier string) int {
	if totalEscrow <= 10000000 {
		switch eoTier {
		case "GOLD":
			return 15
		case "SILVER":
			return 10
		default: // BRONZE
			return 5
		}
	}

	switch eoTier {
	case "GOLD":
		return 8
	case "SILVER":
		return 4
	default: // BRONZE
		return 2
	}
}

// ============================================================================
// 📈 SECTOR 3: MONTHLY LOYALTY TIER EVALUATION (CRM ENGINE) - FULLY SEPARATED
// ============================================================================

// EvaluateMonthlyMerchantTier is the main function that calls line-specific sub-functions
func (c *FinanceCalculator) EvaluateMonthlyMerchantTier(rekberType domain.RekberType, currentLoyalty domain.CRMLoyalty, currentRating float64) (string, string) {
	if currentRating < 4.5 {
		return "BRONZE", "RATING_DROP"
	}

	switch rekberType {
	case domain.TypeEvents:
		return c.evaluateEventLoyalty(currentLoyalty, currentRating)
	case domain.TypeServices:
		return c.evaluateServiceLoyalty(currentLoyalty, currentRating)
	default:
		return "BRONZE", "STAY_BRONZE"
	}
}

// Pure isolation function for evaluating the Event Organizer tier
func (c *FinanceCalculator) evaluateEventLoyalty(loyalty domain.CRMLoyalty, rating float64) (string, string) {
	if rating >= 4.7 && loyalty.TotalCompletedEvents >= 10 {
		return "GOLD", "STAY_GOLD"
	}
	if rating >= 4.5 && loyalty.TotalCompletedEvents >= 3 {
		return "SILVER", "UPGRADE_TO_SILVER"
	}
	return "BRONZE", "STAY_BRONZE"
}

// Pure isolation function for evaluating the Services Provider / Vendor tier
func (c *FinanceCalculator) evaluateServiceLoyalty(loyalty domain.CRMLoyalty, rating float64) (string, string) {
	if rating >= 4.7 && loyalty.TotalCompletedServices >= 20 {
		if loyalty.CurrentTier == "GOLD" && loyalty.ConsecutiveFailedMonths > 0 {
			if loyalty.ConsecutiveFailedMonths >= 3 {
				return "SILVER", "DOWNGRADE_TO_SILVER"
			}
			return "GOLD", "WARNING_LOW_SALES"
		}
		return "GOLD", "STAY_GOLD"
	}
	if rating >= 4.5 && loyalty.TotalCompletedServices >= 5 {
		return "SILVER", "UPGRADE_TO_SILVER"
	}
	return "BRONZE", "STAY_BRONZE"
}
