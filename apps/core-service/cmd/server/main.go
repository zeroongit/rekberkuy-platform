package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rekberkuy/core-service/config"
	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/repository"
	"rekberkuy/core-service/internal/usecase"
	"rekberkuy/core-service/internal/worker"
)

func main() {
	// 1. LOAD CONFIG TERPUSAT (MENGGANTIKAN GODOTENV MANUAL)
	cfg := config.LoadConfig()

	// 2. INISIALISASI DATABASE VIA GORM
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Gagal membuat koneksi database via GORM: %v", err)
	}

	// Set parameter Connection Pool skala industri
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Gagal mengambil instance sql.DB dari GORM: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("🚀 HORE! Backend Go GORM berhasil terhubung dengan aman ke Database Supabase!")

	// 3. AUTOMIGRATE MULTI-PHASE ISOLATED MIGRATION (ANTI-DEADLOCK CONSTRAINT)
	log.Println("Memulai proses AutoMigrate Tahap 1 (Tabel Master & Utama)...")
	err = db.AutoMigrate(
		&domain.UserProfile{},
		&domain.RekberPayWallet{},
		&domain.CRMLoyalty{},
		&domain.VendorCategoryModel{},
		&domain.KYCSubmission{},
		&domain.Transaction{},
		&domain.RekberPayTransaction{},
	)
	if err != nil {
		log.Fatalf("❌ CRITICAL: Proses AutoMigrate Tahap 1 Gagal: %v", err)
	}

	log.Println("Memulai proses AutoMigrate Tahap 2 (Eksekusi Mandiri Terisolasi)...")
	if err := db.AutoMigrate(&domain.Dispute{}); err != nil {
		log.Fatalf("❌ Gagal migrasi Dispute: %v", err)
	}
	if err := db.AutoMigrate(&domain.ServiceMilestone{}); err != nil {
		log.Fatalf("❌ Gagal migrasi ServiceMilestone: %v", err)
	}
	if err := db.AutoMigrate(&domain.TransactionGoods{}); err != nil {
		log.Fatalf("❌ Gagal migrasi TransactionGoods: %v", err)
	}
	if err := db.AutoMigrate(&domain.TransactionServices{}); err != nil {
		log.Fatalf("❌ Gagal migrasi TransactionServices: %v", err)
	}
	if err := db.AutoMigrate(&domain.TransactionEvents{}); err != nil {
		log.Fatalf("❌ Gagal migrasi TransactionEvents: %v", err)
	}
	if err := db.AutoMigrate(&domain.EventOfficialDetails{}); err != nil {
		log.Fatalf("❌ Gagal migrasi EventOfficialDetails: %v", err)
	}
	if err := db.AutoMigrate(&domain.EventVendorPayout{}); err != nil {
		log.Fatalf("❌ Gagal migrasi EventVendorPayout: %v", err)
	}
	if err := db.AutoMigrate(&domain.EventVendorAllocation{}); err != nil {
		log.Fatalf("❌ Gagal migrasi EventVendorAllocation: %v", err)
	}
	if err := db.AutoMigrate(&domain.VendorProfile{}); err != nil {
		log.Fatalf("❌ Gagal migrasi VendorProfile: %v", err)
	}

	log.Println("🎉 SUKSES BULAT! Seluruh 16 tabel terpasang murni tanpa celah di Supabase!")

	// ============================================================================
	// 📦 4. DEPENDENCY INJECTION MAPPING (REPOS, USECASES, HANDLERS)
	// ============================================================================
	
	// Repository Layer
	walletRepo := repository.NewWalletRepository(sqlDB)
	transactionRepo := repository.NewTransactionRepository(sqlDB)
	userRepo := repository.NewUserRepository(sqlDB) // <- SUDAH AKTIF NYATA

	// Usecase Layer
	financeCalc := usecase.NewFinanceCalculator()
	txUsecase := usecase.NewTransactionUsecase(transactionRepo, walletRepo, financeCalc)
	userUsecase := usecase.NewUserUsecase(userRepo, walletRepo) // <- SUDAH AKTIF NYATA

	// Handler Layer
	txHandler := handlers.NewTransactionHandler(txUsecase)
	userHandler := handlers.NewUserHandler(userUsecase) // <- SEKARANG MENERIMA USECASE NYATA

	// ============================================================================
	// 🤖 5. MENYALAKAN BACKGROUND WORKER ROBOT PATROLI
	// ============================================================================
	// Cast repository konkret jika diperlukan, atau passing langsung sesuai tipe
	releaseWorker := worker.NewAutoReleaseWorker(transactionRepo, txUsecase)
	releaseWorker.Start(context.Background())

	// ============================================================================
	// 📡 6. HTTP ROUTING ENGINE & MIDDLEWARE PROTECTION
	// ============================================================================
	r := gin.Default()

	// Pasang tameng CORS untuk membuka gerbang integrasi dengan Next.js
	r.Use(handlers.CORSMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Rekberkuy Engine is running smoothly with Full Security",
		})
	})

	api := r.Group("/api/v1")
	{
		// 🔓 JALUR PUBLIK & SISTEM
		api.POST("/users/register", userHandler.RegisterProfileHandler)
		api.POST("/webhooks/midtrans", txHandler.MidtransWebhookHandler)

		// 🛍️ KLASTER TRANSAKSI BARANG (GOODS)
		goods := api.Group("/transactions/goods")
		{
			goods.POST("", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.LockFundsAwalHandler)
			goods.POST("/release", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.ReleaseFundsHandler)
		}

		// 💼 KLASTER TRANSAKSI JASA (SERVICES - MILESTONE WORK)
		services := api.Group("/transactions/services")
		{
			services.POST("", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.LockFundsAwalHandler)
			services.POST("/release-milestone", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.ReleaseMilestoneHandler)
		}

		// 🎪 KLASTER TRANSAKSI ACARA (EVENTS - MULTI VENDOR)
		events := api.Group("/transactions/events")
		{
			events.POST("", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.LockFundsAwalHandler)
			events.POST("/release-milestone", handlers.AuthRoleMiddleware(domain.RoleAdmin), txHandler.ReleaseEventMilestoneHandler)
			events.POST("/release-vendors", handlers.AuthRoleMiddleware(domain.RoleEventOrganizer, domain.RoleAdmin), txHandler.ProcessEventVendorPayoutHandler)
		}
	}

	log.Printf("Server berjalan di port %s...", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}