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
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Gagal membuat koneksi database via GORM: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Gagal mengambil instance sql.DB dari GORM: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("🚀 HORE! Backend Go GORM berhasil terhubung dengan aman ke Database Supabase!")

	log.Println("Memulai proses AutoMigrate Tahap 1 (Tabel Klasifikasi, Master & Utama)...")
	err = db.AutoMigrate(
		&domain.GoodsCategory{},
		&domain.GoodsSubCategory{},
		&domain.GoodsSubSubCategory{},
		&domain.ServiceCategory{},
		&domain.ServiceSubCategory{},
		&domain.ServiceSubSubCategory{},
		&domain.EventCategory{},
		&domain.EventSubCategory{},
		&domain.EventSubSubCategory{},
		&domain.VendorCategoryModel{},
		&domain.UserProfile{},
		&domain.RekberPayWallet{},
		&domain.PlatformFinance{},
		&domain.IdempotencyRecord{},
		&domain.CRMLoyalty{},
		&domain.KYCSubmission{},
		&domain.Transaction{},
		&domain.RekberPayTransaction{},
	)
	if err != nil {
		log.Fatalf("❌ CRITICAL: Proses AutoMigrate Tahap 1 Gagal: %v", err)
	}

	log.Println("Memulai proses AutoMigrate Tahap 2 (Tabel Eksekusi Relasional Terisolasi)...")
	if err := db.AutoMigrate(&domain.Dispute{}); err != nil {
		log.Fatalf("❌ Gagal migrasi Dispute: %v", err)
	}
	if err := db.AutoMigrate(&domain.VendorProfile{}); err != nil {
		log.Fatalf("❌ Gagal migrasi VendorProfile: %v", err)
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

	log.Println("🎉 SUKSES BULAT! Seluruh tabel terpasang murni secara modular di Supabase!")

	// ============================================================================
	// 📦 DEPENDENCY INJECTION MAPPING (REPOS, USECASES, HANDLERS)
	// ============================================================================
	
	// 1. Repository Layer
	walletRepo := repository.NewWalletRepository(sqlDB)
	transactionRepo := repository.NewTransactionRepository(sqlDB)
	userRepo := repository.NewUserRepository(sqlDB) 
	financeRepo := repository.NewFinanceRepository(sqlDB)
	idemRepo := repository.NewIdempotencyRepository(sqlDB)
	kycRepo := repository.NewKYCRepository(sqlDB)
	vendorRepo := repository.NewVendorRepository(sqlDB)

	// 2. Usecase Layer[cite: 11]
	financeCalc := usecase.NewFinanceCalculator()
	userUsecase := usecase.NewUserUsecase(userRepo, walletRepo)
	kycUsecase := usecase.NewKYCUsecase(kycRepo)
	vendorUsecase := usecase.NewVendorUsecase(vendorRepo)

	// Pemecahan Usecase Transaksi Baru
	goodsUsecase := usecase.NewTransactionGoodsUsecase(transactionRepo, walletRepo, financeRepo, financeCalc)
	servicesUsecase := usecase.NewTransactionServicesUsecase(transactionRepo, walletRepo, financeRepo, financeCalc)
	eventsUsecase := usecase.NewTransactionEventsUsecase(transactionRepo, walletRepo, financeRepo, financeCalc)

	// 3. Handler Layer[cite: 11]
	userHandler := handlers.NewUserHandler(userUsecase) 
	walletHandler := handlers.NewWalletHandler(userUsecase)
	kycHandler := handlers.NewKYCHandler(kycUsecase)
	vendorHandler := handlers.NewVendorHandler(vendorUsecase)

	// Pemecahan Handler Transaksi Baru
	goodsHandler := handlers.NewTransactionGoodsHandler(goodsUsecase)
	servicesHandler := handlers.NewTransactionServicesHandler(servicesUsecase)
	eventsHandler := handlers.NewTransactionEventsHandler(eventsUsecase)

	// ============================================================================
	// 🤖 MENYALAKAN BACKGROUND WORKER ROBOT PATROLI[cite: 11]
	// ============================================================================
	releaseWorker := worker.NewAutoReleaseWorker(transactionRepo, goodsUsecase)
	releaseWorker.Start(context.Background())

	// ============================================================================
	// 📡 HTTP ROUTING ENGINE & MIDDLEWARE PROTECTION[cite: 11]
	// ============================================================================
	r := gin.Default()

	r.Use(handlers.CORSMiddleware())
	r.Use(handlers.IdempotencyMiddleware(idemRepo))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Rekberkuy Engine is running smoothly with Multi-Tenant Architecture",
		})
	})

	api := r.Group("/api/v1")
	{
		// 🔓 JALUR PUBLIK & SISTEM INTERNASIONAL[cite: 11]
		api.POST("/users/register", userHandler.RegisterProfileHandler)
		api.POST("/users/token-test", userHandler.GenerateTokenTestHandler)
		
		// Dokumen KYC & Registrasi Entitas Mitra
		api.POST("/kyc/submit", handlers.AuthRoleMiddleware(domain.RoleUser), kycHandler.SubmitKYCHandler)
		api.POST("/vendors/register", handlers.AuthRoleMiddleware(domain.RoleUser), vendorHandler.RegisterVendorHandler)

		// 🛍️ KLASTER TRANSAKSI BARANG (GOODS GROUP)[cite: 11]
		goodsGroup := api.Group("/transactions/goods")
		{
			goodsGroup.POST("/lock", handlers.AuthRoleMiddleware(domain.RoleUser), goodsHandler.LockFundsGoodsHandler)
			goodsGroup.POST("/release", handlers.AuthRoleMiddleware(domain.RoleUser), goodsHandler.ReleaseGoodsHandler)
		}

		// 💼 KLASTER TRANSAKSI JASA (SERVICES GROUP)[cite: 11]
		servicesGroup := api.Group("/transactions/services")
		{
			servicesGroup.POST("/lock", handlers.AuthRoleMiddleware(domain.RoleUser), servicesHandler.LockFundsServicesHandler)
			servicesGroup.POST("/release-milestone", handlers.AuthRoleMiddleware(domain.RoleUser), servicesHandler.ReleaseMilestoneHandler)
		}

		// 🎪 KLASTER TRANSAKSI ACARA (EVENTS GROUP)[cite: 11]
		eventsGroup := api.Group("/transactions/events")
		{
			eventsGroup.POST("/lock", handlers.AuthRoleMiddleware(domain.RoleUser), eventsHandler.LockFundsEventsHandler)
			eventsGroup.POST("/release-milestone", handlers.AuthRoleMiddleware(domain.RoleAdmin), eventsHandler.ReleaseEventMilestoneHandler)
			eventsGroup.POST("/release-vendors", handlers.AuthRoleMiddleware(domain.RoleEventOrganizer, domain.RoleAdmin), eventsHandler.ProcessEventVendorPayoutHandler)
		}

		// 💳 KLASTER GERBANG LOG PEMBAYARAN WALLET[cite: 11]
		wallets := api.Group("/wallets")
		{
			wallets.POST("/topup", handlers.AuthRoleMiddleware(domain.RoleUser), walletHandler.CreateTopUpHandler)
		}
	}

	log.Printf("Server berjalan di port %s...", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}