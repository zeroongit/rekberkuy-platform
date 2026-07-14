package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rekberkuy/core-service/config"
	"rekberkuy/core-service/internal/domain"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatalf("❌ SEEDER: Gagal terhubung ke database Supabase: %v", err)
	}

	fmt.Println("🚀 SEEDER: Berhasil terhubung! Mulai menyuntikkan Data Master Kategori & Mockup UUID Transaksi...")

	// ============================================================================
	// 🛍️ 1. SEEDER: KATEGORI 3-TIER (GOODS, SERVICES, EVENTS, VENDORS)
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
	fmt.Println("✅ SEEDER: Seluruh struktur Kategori 3-Tier sinkron.")

	// ============================================================================
	// 👥 2. MOCKUP DATA: USER PROFILES (Format ID Wajib Valid UUID)
	// ============================================================================
	buyerID := "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"
	sellerID := "b2c3d4e5-f6a7-8b9c-0d1e-2f3a4b5c6d7e"

	users := []domain.UserProfile{
		{ID: buyerID, Username: "habibullah_buyer", FullName: "Habibullah Buyer", Role: domain.RoleUser, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: sellerID, Username: "karis_seller", FullName: "Karis Verified Merchant", Role: domain.RoleVerifiedMerchant, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for i := range users {
		db.Where(domain.UserProfile{Username: users[i].Username}).FirstOrCreate(&users[i])
	}
	fmt.Println("✅ SEEDER: Mockup User Profiles berbasis UUID sukses disuntik.")

	// ============================================================================
	// 💳 3. MOCKUP DATA: WALLETS REKBERPAY
	// ============================================================================
	for _, u := range users {
		wallet := domain.RekberPayWallet{
			UserID:    u.ID,
			Balance:   10000000, // Rp 10.000.000
			IsFrozen:  false,
			UpdatedAt: time.Now(),
		}
		db.Where(domain.RekberPayWallet{UserID: u.ID}).FirstOrCreate(&wallet)
	}
	fmt.Println("✅ SEEDER: Mockup Wallet RekberPay dengan saldo terpasang.")

	// ============================================================================
	// 🧾 4. MOCKUP DATA: TRANSAKSI DUMMY & DETAIL TRANSAKSI GOODS (SINKRON 100%)
	// ============================================================================
	txID := "9f8e7d6c-5b4a-3f2e-1d0c-9b8a7f6e5d4c"

	dummyTx := domain.Transaction{
		ID:              txID,
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Type:            domain.TypeGoods,
		Status:          domain.StatusFundsLocked, // Status terkunci agar bisa dipatroli worker
		AmountBase:      1500000,
		ShippingFee:     25000,
		ServiceFee:      domain.FeeGoodsNonRekberPay,
		MidtransFee:     4000,
		AmountGross:     1529000, // AmountBase + ShippingFee + ServiceFee + MidtransFee
		AmountNet:       1500000,
		MidtransOrderID: "MID-SANDBOX-MOCK-001",
		IdempotencyKey:  "IDEM-KEY-MOCK-GOODS-001",
		PaymentMethod:   "BANK_TRANSFER_PERMATA",
		CreatedAt:       time.Now().Add(-73 * time.Hour), // Dibuat > 3 hari lalu untuk memicu patroli auto-confirm
		UpdatedAt:       time.Now().Add(-73 * time.Hour),
	}
	db.Where(domain.Transaction{ID: dummyTx.ID}).FirstOrCreate(&dummyTx)

	// Pastikan detail penampung barang sinkron menggunakan foreign key sub-sub category
	dummyGoods := domain.TransactionGoods{
		TransactionID:          txID,
		SubSubCategoryID:      goodsSubSub.ID,
		ShippingCourier:        "JNE OKE",
		ShippingAddress:        "Jakarta Barat, DKI Jakarta",
		AutoConfirmDeadline:    time.Now().Add(-1 * time.Hour), // Sudah melewati batas waktu konfirmasi
	}
	db.Where(domain.TransactionGoods{TransactionID: dummyGoods.TransactionID}).FirstOrCreate(&dummyGoods)

	fmt.Println("🎉 MANTAP TOTAL! Seluruh data mockup terkalibrasi pas dengan logika struct asli dan siap digunakan untuk uji fungsi!")
}