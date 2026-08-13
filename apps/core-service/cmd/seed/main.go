package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rekberkuy/core-service/config"
	"rekberkuy/core-service/internal/domain"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatalf("❌ SEEDER: Failed to connect to Supabase database: %v", err)
	}

	fmt.Println("🚀 SEEDER: Successfully connected! Starting to inject Category Master Data & Mockup Transaction UUIDs...")

	// ============================================================================
	// 🛍️ 1. SEEDER: 3-TIER CATEGORIES (GOODS, SERVICES, EVENTS, VENDORS)
	// ============================================================================
	goodsCat := domain.GoodsCategory{Name: "ELEKTRONIK", Slug: "elektronik", CreatedAt: time.Now()}
	db.Where(domain.GoodsCategory{Slug: "elektronik"}).FirstOrCreate(&goodsCat)

	goodsSub := domain.GoodsSubCategory{CategoryID: goodsCat.ID, Name: "GADGET", Slug: "gadget", CreatedAt: time.Now()}
	db.Where(domain.GoodsSubCategory{Slug: "gadget"}).FirstOrCreate(&goodsSub)

	goodsSubSub := domain.GoodsSubSubCategory{SubCategoryID: goodsSub.ID, Name: "Smartphone Android", Slug: "smartphone-android", CreatedAt: time.Now()}
	db.Where(domain.GoodsSubSubCategory{Slug: "smartphone-android"}).FirstOrCreate(&goodsSubSub)

	srvCat := domain.ServiceCategory{Name: "TEKNOLOGI", Slug: "teknologi", CreatedAt: time.Now()}
	db.Where(domain.ServiceCategory{Slug: "teknologi"}).FirstOrCreate(&srvCat)

	srvSub := domain.ServiceSubCategory{CategoryID: srvCat.ID, Name: "SOFTWARE_DEVELOPMENT", Slug: "software-development", CreatedAt: time.Now()}
	db.Where(domain.ServiceSubCategory{Slug: "software-development"}).FirstOrCreate(&srvSub)

	srvSubSub := domain.ServiceSubSubCategory{SubCategoryID: srvSub.ID, Name: "Backend Golang Development", Slug: "backend-golang-development", CreatedAt: time.Now()}
	db.Where(domain.ServiceSubSubCategory{Slug: "backend-golang-development"}).FirstOrCreate(&srvSubSub)

	vendorCat := domain.VendorCategoryModel{Name: "VENDOR", Slug: "vendor", CreatedAt: time.Now()}
	db.Where(domain.VendorCategoryModel{Slug: "vendor"}).FirstOrCreate(&vendorCat)

	soundSub := domain.VendorSubCategory{CategoryID: vendorCat.ID, Name: "SOUND_SYSTEM", Slug: "sound-system", CreatedAt: time.Now()}
	db.Where(domain.VendorSubCategory{Slug: "sound-system"}).FirstOrCreate(&soundSub)

	soundSubSub := domain.VendorSubSubCategory{SubCategoryID: soundSub.ID, Name: "Line Array System", Slug: "line-array-system", CreatedAt: time.Now()}
	db.Where(domain.VendorSubSubCategory{Slug: "line-array-system"}).FirstOrCreate(&soundSubSub)
	fmt.Println("✅ SEEDER: All 3-Tier Category structures synced.")

	// ============================================================================
	// 👥 2. MOCKUP DATA: USER PROFILES (ID Format Must Be Valid UUID)
	// ============================================================================
	buyerID := "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"
	sellerID := "b2c3d4e5-f6a7-8b9c-0d1e-2f3a4b5c6d7e"

	buyerHash, err := bcrypt.GenerateFromPassword([]byte("buyer123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ SEEDER: failed to hash buyer password: %v", err)
	}
	sellerHash, err := bcrypt.GenerateFromPassword([]byte("seller123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ SEEDER: failed to hash seller password: %v", err)
	}

	users := []domain.UserProfile{
		{ID: buyerID, Username: "habibullah_buyer", Email: "buyer@rekberkuy.id", PasswordHash: string(buyerHash), FullName: "Habibullah Buyer", Role: domain.RoleUser, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: sellerID, Username: "karis_seller", Email: "seller@rekberkuy.id", PasswordHash: string(sellerHash), FullName: "Karis Verified Merchant", Role: domain.RoleVerifiedMerchant, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for i := range users {
		db.Where(domain.UserProfile{Username: users[i].Username}).FirstOrCreate(&users[i])
	}
	fmt.Println("✅ SEEDER: UUID-based Mockup User Profiles successfully injected.")

	// ============================================================================
	// 💳 3. MOCKUP DATA: REKBERPAY WALLETS
	// ============================================================================
	for _, u := range users {
		wallet := domain.RekberPayWallet{
			UserID:    u.ID,
			Balance:   10000000, // IDR 10,000,000
			IsFrozen:  false,
			UpdatedAt: time.Now(),
		}
		db.Where(domain.RekberPayWallet{UserID: u.ID}).FirstOrCreate(&wallet)
	}
	fmt.Println("✅ SEEDER: Mockup RekberPay Wallets with balance installed.")

	// ============================================================================
	// 🧾 4. MOCKUP DATA: DUMMY TRANSACTIONS & GOODS TRANSACTION DETAILS (100% SYNCED)
	// ============================================================================
	txID := "9f8e7d6c-5b4a-3f2e-1d0c-9b8a7f6e5d4c"

	dummyTx := domain.Transaction{
		ID:              txID,
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Type:            domain.TypeGoods,
		Status:          domain.StatusFundsLocked, // Locked status so it can be patrolled by the worker
		AmountBase:      1500000,
		ShippingFee:     25000,
		ServiceFee:      domain.FeeGoodsNonRekberPay,
		MidtransFee:     4000,
		AmountGross:     1529000, // AmountBase + ShippingFee + ServiceFee + MidtransFee
		AmountNet:       1500000,
		MidtransOrderID: "MID-SANDBOX-MOCK-001",
		IdempotencyKey:  "IDEM-KEY-MOCK-GOODS-001",
		PaymentMethod:   "BANK_TRANSFER_PERMATA",
		CreatedAt:       time.Now().Add(-73 * time.Hour), // Created > 3 days ago to trigger auto-confirm patrol
		UpdatedAt:       time.Now().Add(-73 * time.Hour),
	}
	db.Where(domain.Transaction{ID: dummyTx.ID}).FirstOrCreate(&dummyTx)

	// Ensure the goods detail container is synced using the sub-sub category foreign key
	dummyGoods := domain.TransactionGoods{
		TransactionID:       txID,
		SubSubCategoryID:    goodsSubSub.ID,
		ShippingCourier:     "JNE OKE",
		ShippingAddress:     "Jakarta Barat, DKI Jakarta",
		AutoConfirmDeadline: time.Now().Add(-1 * time.Hour), // Past the confirmation deadline
	}
	db.Where(domain.TransactionGoods{TransactionID: dummyGoods.TransactionID}).FirstOrCreate(&dummyGoods)

	fmt.Println("🎉 PERFECT! All mockup data is properly calibrated with the original struct logic and ready for testing!")
}
