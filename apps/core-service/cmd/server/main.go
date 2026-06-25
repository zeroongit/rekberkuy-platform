package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/joho/godotenv"
	"rekberkuy/core-service/internal/domain"
)

func main() {
	// 1. Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: Gagal memuat file .env, sistem akan membaca OS environment variables")
	}

	// 2. Ambil string koneksi database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Error: DATABASE_URL tidak ditemukan di file .env")
	}

	// 3. Inisialisasi Database via GORM (Mengelola Connection Pool secara otomatis)
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Menampilkan log query SQL di terminal
	})
	if err != nil {
		log.Fatalf("Gagal membuat koneksi database via GORM: %v", err)
	}

	// Set parameter Connection Pool skala industri
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	fmt.Println("🚀 HORE! Backend Go GORM berhasil terhubung dengan aman ke Database Supabase!")

	// 4. EKSEKUSI AUTOMIGRATE SEBELUM API SERVER NYALA
	log.Println("Memulai proses AutoMigrate lintas entitas...")
	err = db.AutoMigrate(
		&domain.UserProfile{},
		&domain.RekberPayWallet{},
		&domain.CRMLoyalty{},
		&domain.Transaction{},
		&domain.TransactionGoods{},
		&domain.TransactionServices{},
		&domain.TransactionEvents{},
		&domain.ServiceMilestone{},
		&domain.VendorCategoryModel{},
		&domain.VendorProfile{},
		&domain.EventVendorAllocation{},
		&domain.EventOfficialDetails{},
		&domain.EventVendorPayout{},
		&domain.Dispute{},
		&domain.RekberPayTransaction{},
		&domain.KYCSubmission{},
	)
	if err != nil {
		log.Fatalf("❌ CRITICAL: Proses AutoMigrate GORM Gagal: %v", err)
	}
	log.Println("🎉 SUKSES! Seluruh tabel dan foreign key terpasang rinci di Supabase!")

	// 5. Inisialisasi Server Gin HTTP
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Rekberkuy Engine is running smoothly",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di port %s...", port)
	r.Run(":" + port)
}