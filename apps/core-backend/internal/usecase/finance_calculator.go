package usecase

import (
	"rekberkuy/core-backend/internal/domain"
)

type FinanceCalculator struct{}

func NewFinanceCalculator() *FinanceCalculator {
	return &FinanceCalculator{}
}

func (c *FinanceCalculator) CalculateBuyerServiceFee(rekberType domain.RekberType, amountBase int64, isRekberPay bool, sellerTier string) int64 {
	if rekberType == domain.TypeGoods || rekberType == domain.TypeServices {
		switch sellerTier {
		case "GOLD":
			return int64(float64(amountBase) * 0.08)
		case "SILVER":
			return int64(float64(amountBase) * 0.04)
		default:
			if isRekberPay {
				return domain.FeeGoodsRekberPay
			}
			return domain.FeeGoodsNonRekberPay
		}
	}
	return 0
}

func (c *FinanceCalculator) CalculateSellerServiceFee(rekberType domain.RekberType, amountToRelease int64, sellerTier string) int64 {
	if rekberType == domain.TypeGoods || rekberType == domain.TypeServices {
		switch sellerTier {
		case "GOLD":
			return int64(float64(amountToRelease) * 0.03)
		case "SILVER":
			return int64(float64(amountToRelease) * 0.06)
		default:
			return int64(float64(amountToRelease) * 0.10)
		}
	}
	return 0
}

func (c *FinanceCalculator) CalculateEventAudit(totalEscrowLocked int64, vendorsBill []domain.EventVendorAllocation, eoTier string) domain.EventAuditResult {
	
	// 1. Hitung total pengeluaran riil di lapangan (Sum seluruh tagihan akhir vendor)
	var totalActualSpent int64 = 0
	for _, bill := range vendorsBill {
		totalActualSpent += bill.ActualPaidAmount
	}

	// 2. Hitung Fee Platform Rekberkuy (Flat 5% dari dana yang benar-benar terpakai oleh para vendor)
	var platformFee int64 = 0
	if totalActualSpent > 0 {
		platformFee = int64(float64(totalActualSpent) * 0.05)
	}

	// 3. Hitung sisa anggaran bersih (surplus) setelah dipotong biaya vendor & komisi platform kita
	rawSurplus := totalEscrowLocked - totalActualSpent - platformFee
	if rawSurplus < 0 {
		rawSurplus = 0
	}

	// 4. Menentukan persentase bonus insentif EO dari sisa anggaran bersih (rawSurplus)
	var eoBonusPercent float64 = 0.02 // Default dasar untuk Event Besar (> Rp 10 Juta)

	if totalEscrowLocked <= 10000000 {
		// === KLASTER EVENT KECIL (<= Rp 10 Juta) ===
		switch eoTier {
		case "GOLD":
			eoBonusPercent = 0.15 // 15% murni untuk EO Gold
		case "SILVER":
			eoBonusPercent = 0.10 // 10% untuk Silver
		default:
			eoBonusPercent = 0.05 // 5% untuk Newbie
		}
	} else {
		// === KLASTER EVENT BESAR (> Rp 10 Juta) ===
		switch eoTier {
		case "GOLD":
			eoBonusPercent = 0.08 // 8%
		case "SILVER":
			eoBonusPercent = 0.04 // 4%
		default:
			eoBonusPercent = 0.02 // 2%
		}
	}

	// 5. Hitung nominal rupiah final untuk jatah bonus EO dan sisa pengembalian otomatis ke peserta
	bonusToEO := int64(float64(rawSurplus) * eoBonusPercent)
	refundToPeserta := rawSurplus - bonusToEO
	if refundToPeserta < 0 {
		refundToPeserta = 0
	}

	return domain.EventAuditResult{
		PlatformFee:     platformFee,
		AmountToVendor:  totalActualSpent, // Total dana yang didistribusikan ke seluruh vendor
		BonusToEO:       bonusToEO,
		RefundToPeserta: refundToPeserta,
	}
}

func (c *FinanceCalculator) EvaluateMonthlyMerchantTier(rekberType domain.RekberType, currentLoyalty domain.CRMLoyalty, currentRating float64) (string, string) {
	if currentRating < 4.5 {
		return "NEWBIE", "RATING_DROP"
	}

	if rekberType == domain.TypeEvents {
		if currentRating >= 4.7 && currentLoyalty.TotalCompletedEvents >= 10 {
			return "GOLD", "STAY_GOLD"
		}
		if currentRating >= 4.5 && currentLoyalty.TotalCompletedEvents >= 3 {
			return "SILVER", "UPGRADE_TO_SILVER"
		}
		return "NEWBIE", "STAY_NEWBIE"
	}

	if rekberType == domain.TypeServices {
		if currentRating >= 4.7 && currentLoyalty.TotalCompletedServices >= 20 {
			if currentLoyalty.CurrentTier == "GOLD" && currentLoyalty.ConsecutiveFailedMonths > 0 {
				if currentLoyalty.ConsecutiveFailedMonths >= 3 {
					return "SILVER", "DOWNGRADE_TO_SILVER"
				}
				return "GOLD", "WARNING_LOW_SALES"
			}
			return "GOLD", "STAY_GOLD"
		}
		if currentRating >= 4.5 && currentLoyalty.TotalCompletedServices >= 5 {
			return "SILVER", "UPGRADE_TO_SILVER"
		}
		return "NEWBIE", "STAY_NEWBIE"
	}

	var monthlyTarget int64 = 100000000
	var silverTotalRequired int64 = 100000000
	var goldTotalRequired int64 = 1000000000

	if currentLoyalty.MaxItemPriceSold <= 100000 {
		monthlyTarget = 10000000
		silverTotalRequired = 10000000
		goldTotalRequired = 100000000
	} else if currentLoyalty.MaxItemPriceSold > 10000000 {
		monthlyTarget = 1000000000
		silverTotalRequired = 1000000000
		goldTotalRequired = 10000000000
	}

	if currentLoyalty.CurrentTier == "GOLD" {
		if currentLoyalty.CurrentMonthGmv < monthlyTarget {
			failedMonths := currentLoyalty.ConsecutiveFailedMonths + 1
			if failedMonths >= 3 {
				return "SILVER", "DOWNGRADE_TO_SILVER"
			}
			return "GOLD", "WARNING_LOW_SALES"
		}
		return "GOLD", "STAY_GOLD"
	}

	if currentLoyalty.TotalSpentFiat >= goldTotalRequired {
		return "GOLD", "UPGRADE_TO_GOLD"
	} else if currentLoyalty.TotalSpentFiat >= silverTotalRequired {
		return "SILVER", "UPGRADE_TO_SILVER"
	}

	return "NEWBIE", "STAY_NEWBIE"
}