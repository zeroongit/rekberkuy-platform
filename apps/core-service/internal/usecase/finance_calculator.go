package usecase

import (
	"rekberkuy/core-service/internal/domain" // Sesuaikan path go.mod Anda
)

type FinanceCalculator struct{}

func NewFinanceCalculator() *FinanceCalculator {
	return &FinanceCalculator{}
}

// ============================================================================
// 📦 SEKTOR 1: KALKULASI FEE TRANSAKSI (UPFRONT & RELEASE)
// ============================================================================

// CalculateBuyerServiceFee menghitung biaya proteksi pembeli di awal transaksi
func (c *FinanceCalculator) CalculateBuyerServiceFee(rekberType domain.RekberType, amountBase int64, isRekberPay bool, sellerTier string) int64 {
	if rekberType == domain.TypeServices {
		return 0 
	}

	if rekberType == domain.TypeGoods {
		switch sellerTier {
		case "GOLD":
			return int64(float64(amountBase) * 0.08)
		case "SILVER":
			return int64(float64(amountBase) * 0.04)
		default: // BRONZE
			if isRekberPay {
				return domain.FeeGoodsRekberPay
			}
			return domain.FeeGoodsNonRekberPay
		}
	}
	return 0
}

// CalculateSellerServiceFee menghitung potongan komisi merchant saat dana cair
func (c *FinanceCalculator) CalculateSellerServiceFee(rekberType domain.RekberType, amountToRelease int64, sellerTier string) int64 {
	if rekberType == domain.TypeServices {
		return int64(float64(amountToRelease) * 0.02)
	}

	if rekberType == domain.TypeGoods {
		switch sellerTier {
		case "GOLD":
			return int64(float64(amountToRelease) * 0.03) 
		case "SILVER":
			return int64(float64(amountToRelease) * 0.06) 
		default: // BRONZE
			return int64(float64(amountToRelease) * 0.10) 
		}
	}
	return 0
}

// ============================================================================
// 🎪 SEKTOR 2: AUDIT FINANSIAL LINI EVENT (POST-EVENT AUDIT ENGINE)
// ============================================================================

// CalculateEventAudit menghitung pembagian surplus dana milik Event Organizer
func (c *FinanceCalculator) CalculateEventAudit(totalEscrowLocked int64, vendorsBill []domain.EventVendorAllocation, eoTier string) domain.EventAuditResult {
	platformFee := int64(float64(totalEscrowLocked) * 0.05)
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

	// Evaluasi ambang batas Rp500.000 khusus untuk Event Kecil (<= Rp10 Juta)
	if netSurplus <= 500000 && totalEscrowLocked <= 10000000 {
		bonusToEO = netSurplus
		refundToPeserta = 0
	} else {
		// Panggil fungsi terpisah untuk mendapatkan persentase bonus berdasarkan kasta
		eoBonusPercent := c.getEventBonusPercentage(totalEscrowLocked, eoTier)
		bonusToEO = int64(float64(netSurplus) * eoBonusPercent)
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

// Helper internal untuk memisahkan logika persentase bonus event
func (c *FinanceCalculator) getEventBonusPercentage(totalEscrow int64, eoTier string) float64 {
	if totalEscrow <= 10000000 {
		switch eoTier {
		case "GOLD":
			return 0.15
		case "SILVER":
			return 0.10
		default: // BRONZE
			return 0.05
		}
	}
	
	switch eoTier {
	case "GOLD":
		return 0.08
	case "SILVER":
		return 0.04
	default: // BRONZE
		return 0.02
	}
}

// ============================================================================
// 📈 SEKTOR 3: EVALUASI KASTA LOYALITAS BULANAN (CRM ENGINE) - DIPISAH TOTAL
// ============================================================================

// EvaluateMonthlyMerchantTier adalah fungsi utama yang memanggil sub-fungsi spesifik tiap lini
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

// Fungsi isolasi murni untuk evaluasi kasta Event Organizer
func (c *FinanceCalculator) evaluateEventLoyalty(loyalty domain.CRMLoyalty, rating float64) (string, string) {
	if rating >= 4.7 && loyalty.TotalCompletedEvents >= 10 {
		return "GOLD", "STAY_GOLD"
	}
	if rating >= 4.5 && loyalty.TotalCompletedEvents >= 3 {
		return "SILVER", "UPGRADE_TO_SILVER"
	}
	return "BRONZE", "STAY_BRONZE"
}

// Fungsi isolasi murni untuk evaluasi kasta Penyedia Jasa / Vendor
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