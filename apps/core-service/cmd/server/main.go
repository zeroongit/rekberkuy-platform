package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/joho/godotenv"
	
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/repository" 
	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/usecase"      
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
	if err != nil {
		log.Fatalf("Gagal mengambil instance sql.DB dari GORM: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("🚀 HORE! Backend Go GORM berhasil terhubung dengan aman ke Database Supabase!")

	// 4. EKSEKUSI AUTOMIGRATE LINTAS ENTITAS (STRATEGI SINGLE-ISOLATED STEP MIGRATION)
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

	// Step A: Buat entitas fundamental sengketa dan milestone
	if err := db.AutoMigrate(&domain.Dispute{}); err != nil {
		log.Fatalf("❌ Gagal migrasi Dispute: %v", err)
	}
	if err := db.AutoMigrate(&domain.ServiceMilestone{}); err != nil {
		log.Fatalf("❌ Gagal migrasi ServiceMilestone: %v", err)
	}

	// Step B: Buat tabel spesifikasi lini dasar Goods & Services
	if err := db.AutoMigrate(&domain.TransactionGoods{}); err != nil {
		log.Fatalf("❌ Gagal migrasi TransactionGoods: %v", err)
	}
	if err := db.AutoMigrate(&domain.TransactionServices{}); err != nil {
		log.Fatalf("❌ Gagal migrasi TransactionServices: %v", err)
	}

	// Step C: Lahirkan tabel spesifikasi Event utama agar payout & allocation punya target rujukan
	if err := db.AutoMigrate(&domain.TransactionEvents{}); err != nil {
		log.Fatalf("❌ Gagal migrasi TransactionEvents: %v", err)
	}

	// Step D: Sekarang buat tabel-tabel anak dari event yang tadinya saling mengunci
	if err := db.AutoMigrate(&domain.EventOfficialDetails{}); err != nil {
		log.Fatalf("❌ Gagal migrasi EventOfficialDetails: %v", err)
	}
	if err := db.AutoMigrate(&domain.EventVendorPayout{}); err != nil {
		log.Fatalf("❌ Gagal migrasi EventVendorPayout: %v", err)
	}
	if err := db.AutoMigrate(&domain.EventVendorAllocation{}); err != nil {
		log.Fatalf("❌ Gagal migrasi EventVendorAllocation: %v", err)
	}

	// Step E: Terakhir, jalankan profil vendor setelah alokasi tercipta
	if err := db.AutoMigrate(&domain.VendorProfile{}); err != nil {
		log.Fatalf("❌ Gagal migrasi VendorProfile: %v", err)
	}

	log.Println("🎉 HORE SUKSES BESAR! Seluruh 16 tabel dan foreign key terpasang murni tanpa celah di Supabase!")

	walletRepo := repository.NewWalletRepository(sqlDB)
	transactionRepo := repository.NewTransactionRepository(sqlDB)


	financeCalc := usecase.NewFinanceCalculator()
	txUsecase := usecase.NewTransactionUsecase(transactionRepo, walletRepo, financeCalc)

	txHandler := handlers.NewTransactionHandler(txUsecase)

	userHandler := handlers.NewUserHandler()
	// Inisialisasi Robot Auto-Release Baru Anda
	// (Casting interface atau oper concrete pointer sesuai inisialisasi Repo Anda)
	// releaseWorker := worker.NewAutoReleaseWorker(transactionRepo, txUsecase)
	// releaseWorker.Start(context.Background())

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Rekberkuy Engine is running smoothly",
		})
	})

	api := r.Group("/api/v1")
	{
		// ============================================================================
		// 🔓 JALUR PUBLIK & SISTEM
		// ============================================================================
		api.POST("/users/register", userHandler.RegisterProfileHandler)
		api.POST("/webhooks/midtrans", txHandler.MidtransWebhookHandler) 

		// ============================================================================
		// 🛍️ KLASTER TRANSAKSI BARANG (GOODS)
		// ============================================================================
		goods := api.Group("/transactions/goods")
		{
			// Siapa saja yang login (RoleUser/Buyer) bisa menginisialisasi transaksi beli barang
			goods.POST("", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.LockFundsAwalHandler)
			
			// HANYA Pembeli (RoleUser) yang berhak mencairkan dana escrow barang setelah paket tiba
			goods.POST("/release", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.ReleaseFundsHandler)
		}

		// ============================================================================
		// 💼 KLASTER TRANSAKSI JASA (SERVICES - MILESTONE WORK)
		// ============================================================================
		services := api.Group("/transactions/services")
		{
			// Pembeli membuat kontrak kerja jasa baru
			services.POST("", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.LockFundsAwalHandler)
			
			// Kritis: HANYA Pembeli (RoleUser) yang boleh me-release dana per termin/milestone jasa!
			// Freelancer tidak boleh mencairkan uangnya sendiri tanpa persetujuan klien
			services.POST("/release-milestone", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.ReleaseMilestoneHandler)
		}

		// ============================================================================
		// 🎪 KLASTER TRANSAKSI ACARA (EVENTS - MULTI VENDOR)
		// ============================================================================
		events := api.Group("/transactions/events")
		{
			// Pembeli/Peserta membeli tiket atau mendanai event awal
			events.POST("", handlers.AuthRoleMiddleware(domain.RoleUser), txHandler.LockFundsAwalHandler)
			
			// Proteksi Berlapis: Pencairan operasional bertahap berdasarkan invoice EO 
			// HANYA bisa disetujui oleh ADMIN atau MEDIATOR resmi platform setelah bukti diperiksa
			events.POST("/release-milestone", handlers.AuthRoleMiddleware(domain.RoleAdmin), txHandler.ReleaseEventMilestoneHandler)
			
			// Pemecahan dana final ke seluruh vendor lapangan (Gedung, Katering, dll)
			// Hanya bisa dieksekusi oleh ADMIN atau EVENT_ORGANIZER setelah acara sukses selesai
			events.POST("/release-vendors", handlers.AuthRoleMiddleware(domain.RoleEventOrganizer, domain.RoleAdmin), txHandler.ProcessEventVendorPayoutHandler)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di port %s...", port)
	r.Run(":" + port)
}