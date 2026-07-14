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
	})
	if err != nil {
		log.Fatalf("❌ SEEDER: Gagal terhubung ke database Supabase: %v", err)
	}

	fmt.Println("🚀 SEEDER: Berhasil terhubung! Menyiapkan data master taksonomi multi-tier...")

	// ============================================================================
	// 🏢 SEEDER: KATEGORI VENDOR & EO (Tingkat 1 - Root Peran Bisnis)
	// ============================================================================
	vendorCategory := domain.VendorCategoryModel{
		Name:      "VENDOR",
		Slug:      "vendor",
		CreatedAt: time.Now(),
	}
	
	// Gunakan FirstOrCreate agar tidak duplikat jika dijalankan ulang
	if err := db.Where(domain.VendorCategoryModel{Slug: "vendor"}).FirstOrCreate(&vendorCategory).Error; err != nil {
		log.Fatalf("❌ SEEDER: Gagal membuat root Vendor Category: %v", err)
	}
	fmt.Println("✅ SEEDER: Root Vendor Category dikunci.")

	// ============================================================================
	// 🏢 SEEDER: SUB-KATEGORI VENDOR (Tingkat 2 - Jenis Spesialisasi Lapangan)
	// ============================================================================
	subCategories := []domain.VendorSubCategory{
		{CategoryID: vendorCategory.ID, Name: "EVENT_ORGANIZER", Slug: "event-organizer", CreatedAt: time.Now()},
		{CategoryID: vendorCategory.ID, Name: "SOUND_SYSTEM", Slug: "sound-system", CreatedAt: time.Now()},
		{CategoryID: vendorCategory.ID, Name: "KATERING", Slug: "katering", CreatedAt: time.Now()},
		{CategoryID: vendorCategory.ID, Name: "GEDUNG", Slug: "gedung", CreatedAt: time.Now()},
		{CategoryID: vendorCategory.ID, Name: "DEKORASI", Slug: "dekorasi", CreatedAt: time.Now()},
	}

	for i := range subCategories {
		if err := db.Where(domain.VendorSubCategory{Slug: subCategories[i].Slug}).FirstOrCreate(&subCategories[i]).Error; err != nil {
			log.Printf("⚠️ SEEDER WARNING: Gagal memproses sub-category %s: %v", subCategories[i].Name, err)
		}
	}
	fmt.Println("✅ SEEDER: Lapis 2 Sub-Kategori Vendor (EO, Sound, Catering, dll) berhasil disuntik.")

	// ============================================================================
	// 🏢 SEEDER: SUB-SUB-KATEGORI VENDOR (Tingkat 3 - Detil Komersial Lapangan)
	// ============================================================================
	
	// Kita ambil ID untuk Sound System dan EO untuk pemetaan akurat
	var soundSub domain.VendorSubCategory
	db.Where("slug = ?", "sound-system").First(&soundSub)

	if soundSub.ID != 0 {
		subSubCategories := []domain.VendorSubSubCategory{
			{SubCategoryID: soundSub.ID, Name: "Line Array System", Slug: "line-array-system", CreatedAt: time.Now()},
			{SubCategoryID: soundSub.ID, Name: "Subwoofer Ground Stack", Slug: "subwoofer-ground-stack", CreatedAt: time.Now()},
			{SubCategoryID: soundSub.ID, Name: "Wireless Microphone Management", Slug: "wireless-microphone-management", CreatedAt: time.Now()},
		}

		for i := range subSubCategories {
			if err := db.Where(domain.VendorSubSubCategory{Slug: subSubCategories[i].Slug}).FirstOrCreate(&subSubCategories[i]).Error; err != nil {
				log.Printf("⚠️ SEEDER WARNING: Gagal memproses sub-sub-category %s: %v", subSubCategories[i].Name, err)
			}
		}
		fmt.Println("✅ SEEDER: Lapis 3 Sub-Sub-Kategori (Spesialisasi Audio Audio Lapangan) siap.")
	}

	fmt.Println("🎉 MANTAP! Seluruh data master taksonomi berhasil hidup di Supabase secara aman!")
}