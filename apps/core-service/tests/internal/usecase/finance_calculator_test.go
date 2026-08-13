package usecase_test

import (
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// FINANCE CALCULATOR — UNIT TESTS (TABLE-DRIVEN)
// ============================================================================
// finance_calculator is pure logic with no external dependencies, so it is
// well suited for exhaustive table-driven testing.

func TestCalculateBuyerServiceFee(t *testing.T) {
	calc := usecase.NewFinanceCalculator()

	tests := []struct {
		name        string
		rekberType  domain.RekberType
		amountBase  int64
		isRekberPay bool
		sellerTier  string
		want        int64
	}{
		// Goods — GOLD 8%
		{"goods gold 8%", domain.TypeGoods, 100000, true, "GOLD", 8000},
		{"goods gold 8% truncated", domain.TypeGoods, 99999, true, "GOLD", 7999},
		// Goods — SILVER 4%
		{"goods silver 4%", domain.TypeGoods, 100000, true, "SILVER", 4000},
		// Goods — BRONZE flat
		{"goods bronze rekberpay flat 2500", domain.TypeGoods, 1000000, true, "BRONZE", domain.FeeGoodsRekberPay},
		{"goods bronze non-rekberpay flat 5000", domain.TypeGoods, 1000000, false, "BRONZE", domain.FeeGoodsNonRekberPay},
		// Services & Events -> buyer fee 0
		{"services always 0", domain.TypeServices, 100000, true, "GOLD", 0},
		{"events always 0", domain.TypeEvents, 100000, false, "BRONZE", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calc.CalculateBuyerServiceFee(tc.rekberType, tc.amountBase, tc.isRekberPay, tc.sellerTier)
			if got != tc.want {
				t.Errorf("CalculateBuyerServiceFee = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCalculateSellerServiceFee(t *testing.T) {
	calc := usecase.NewFinanceCalculator()

	tests := []struct {
		name       string
		rekberType domain.RekberType
		amount     int64
		sellerTier string
		want       int64
	}{
		// Services — flat 5% (tier-independent)
		{"services flat 5%", domain.TypeServices, 100000, "BRONZE", 5000},
		{"services flat 5% ignore tier", domain.TypeServices, 100000, "GOLD", 5000},
		// Goods — tiered
		{"goods gold 3%", domain.TypeGoods, 100000, "GOLD", 3000},
		{"goods silver 6%", domain.TypeGoods, 100000, "SILVER", 6000},
		{"goods bronze 10%", domain.TypeGoods, 100000, "BRONZE", 10000},
		// Events -> 0
		{"events always 0", domain.TypeEvents, 100000, "GOLD", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calc.CalculateSellerServiceFee(tc.rekberType, tc.amount, tc.sellerTier)
			if got != tc.want {
				t.Errorf("CalculateSellerServiceFee = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCalculateEventAudit(t *testing.T) {
	calc := usecase.NewFinanceCalculator()

	// Helper to build a list of vendor allocations with a given ActualPaidAmount.
	makeBills := func(paid ...int64) []domain.EventVendorAllocation {
		bills := make([]domain.EventVendorAllocation, 0, len(paid))
		for _, p := range paid {
			bills = append(bills, domain.EventVendorAllocation{ActualPaidAmount: p})
		}
		return bills
	}

	tests := []struct {
		name              string
		totalEscrowLocked int64
		vendorsBill       []domain.EventVendorAllocation
		eoTier            string
		wantPlatformFee   int64
		wantToVendor      int64
		wantBonusToEO     int64
		wantRefund        int64
	}{
		{
			name:              "event with large surplus (>500k), EO GOLD takes 15%",
			totalEscrowLocked: 10000000,
			vendorsBill:       makeBills(4000000),
			eoTier:            "GOLD",
			wantPlatformFee:   500000,  // 5%
			wantToVendor:      4000000, // actual
			wantBonusToEO:     825000,  // (9.5m-4m)*0.15 = 825k
			wantRefund:        4675000, // 5.5m - 825k
		},
		{
			name:              "small event surplus <= 500k -> entire surplus to EO",
			totalEscrowLocked: 5000000,
			vendorsBill:       makeBills(4400000),
			eoTier:            "BRONZE",
			wantPlatformFee:   250000, // 5%
			wantToVendor:      4400000,
			wantBonusToEO:     350000, // netSurplus 4.75m-4.4m = 350k -> all to EO
			wantRefund:        0,
		},
		{
			name:              "vendor spending exceeds pool -> cap to maxVendorPool",
			totalEscrowLocked: 1000000,
			vendorsBill:       makeBills(2000000),
			eoTier:            "SILVER",
			wantPlatformFee:   50000,  // 5%
			wantToVendor:      950000, // cap: 1m - 50k
			wantBonusToEO:     0,      // netSurplus 0
			wantRefund:        0,
		},
		{
			name:              "large event (>10m), EO GOLD takes 8%",
			totalEscrowLocked: 20000000,
			vendorsBill:       makeBills(10000000),
			eoTier:            "GOLD",
			wantPlatformFee:   1000000, // 5%
			wantToVendor:      10000000,
			wantBonusToEO:     720000,  // (19m-10m)*0.08
			wantRefund:        8280000, // 9m - 720k
		},
		{
			name:              "multi-vendor: actual total accumulated",
			totalEscrowLocked: 10000000,
			vendorsBill:       makeBills(1000000, 2000000, 500000), // total 3.5m
			eoTier:            "SILVER",
			wantPlatformFee:   500000,
			wantToVendor:      3500000,
			wantBonusToEO:     600000,  // netSurplus 6m * 0.10
			wantRefund:        5400000, // 6m - 600k
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calc.CalculateEventAudit(tc.totalEscrowLocked, tc.vendorsBill, tc.eoTier)
			if got.PlatformFee != tc.wantPlatformFee {
				t.Errorf("PlatformFee = %d, want %d", got.PlatformFee, tc.wantPlatformFee)
			}
			if got.AmountToVendor != tc.wantToVendor {
				t.Errorf("AmountToVendor = %d, want %d", got.AmountToVendor, tc.wantToVendor)
			}
			if got.BonusToEO != tc.wantBonusToEO {
				t.Errorf("BonusToEO = %d, want %d", got.BonusToEO, tc.wantBonusToEO)
			}
			if got.RefundToPeserta != tc.wantRefund {
				t.Errorf("RefundToPeserta = %d, want %d", got.RefundToPeserta, tc.wantRefund)
			}
		})
	}
}

func TestEvaluateMonthlyMerchantTier(t *testing.T) {
	calc := usecase.NewFinanceCalculator()

	tests := []struct {
		name          string
		rekberType    domain.RekberType
		loyalty       domain.CRMLoyalty
		rating        float64
		wantTier      string
		wantActionMsg string
	}{
		// Low rating -> universal drop
		{
			name:          "rating < 4.5 -> drop to BRONZE",
			rekberType:    domain.TypeEvents,
			loyalty:       domain.CRMLoyalty{CurrentTier: "GOLD", TotalCompletedEvents: 20},
			rating:        4.2,
			wantTier:      "BRONZE",
			wantActionMsg: "RATING_DROP",
		},
		// Events
		{
			name:          "event GOLD retained (rating>=4.7, events>=10)",
			rekberType:    domain.TypeEvents,
			loyalty:       domain.CRMLoyalty{CurrentTier: "GOLD", TotalCompletedEvents: 10},
			rating:        4.8,
			wantTier:      "GOLD",
			wantActionMsg: "STAY_GOLD",
		},
		{
			name:          "event upgrade SILVER (rating>=4.5, events>=3)",
			rekberType:    domain.TypeEvents,
			loyalty:       domain.CRMLoyalty{CurrentTier: "BRONZE", TotalCompletedEvents: 3},
			rating:        4.6,
			wantTier:      "SILVER",
			wantActionMsg: "UPGRADE_TO_SILVER",
		},
		{
			name:          "event stays BRONZE (events < 3)",
			rekberType:    domain.TypeEvents,
			loyalty:       domain.CRMLoyalty{CurrentTier: "BRONZE", TotalCompletedEvents: 1},
			rating:        4.9,
			wantTier:      "BRONZE",
			wantActionMsg: "STAY_BRONZE",
		},
		// Services
		{
			name:          "service GOLD retained",
			rekberType:    domain.TypeServices,
			loyalty:       domain.CRMLoyalty{CurrentTier: "SILVER", TotalCompletedServices: 25},
			rating:        4.8,
			wantTier:      "GOLD",
			wantActionMsg: "STAY_GOLD",
		},
		{
			name:          "service GOLD warning (failed months 1-2)",
			rekberType:    domain.TypeServices,
			loyalty:       domain.CRMLoyalty{CurrentTier: "GOLD", TotalCompletedServices: 25, ConsecutiveFailedMonths: 1},
			rating:        4.8,
			wantTier:      "GOLD",
			wantActionMsg: "WARNING_LOW_SALES",
		},
		{
			name:          "service GOLD downgrade to SILVER (failed months>=3)",
			rekberType:    domain.TypeServices,
			loyalty:       domain.CRMLoyalty{CurrentTier: "GOLD", TotalCompletedServices: 25, ConsecutiveFailedMonths: 3},
			rating:        4.8,
			wantTier:      "SILVER",
			wantActionMsg: "DOWNGRADE_TO_SILVER",
		},
		{
			name:          "service upgrade SILVER",
			rekberType:    domain.TypeServices,
			loyalty:       domain.CRMLoyalty{CurrentTier: "BRONZE", TotalCompletedServices: 5},
			rating:        4.6,
			wantTier:      "SILVER",
			wantActionMsg: "UPGRADE_TO_SILVER",
		},
		// Goods default
		{
			name:          "goods always BRONZE in this evaluator",
			rekberType:    domain.TypeGoods,
			loyalty:       domain.CRMLoyalty{CurrentTier: "BRONZE"},
			rating:        5.0,
			wantTier:      "BRONZE",
			wantActionMsg: "STAY_BRONZE",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTier, gotMsg := calc.EvaluateMonthlyMerchantTier(tc.rekberType, tc.loyalty, tc.rating)
			if gotTier != tc.wantTier {
				t.Errorf("tier = %q, want %q", gotTier, tc.wantTier)
			}
			if gotMsg != tc.wantActionMsg {
				t.Errorf("actionMsg = %q, want %q", gotMsg, tc.wantActionMsg)
			}
		})
	}
}
