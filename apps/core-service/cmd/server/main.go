package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rekberkuy/core-service/config"
	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/disbursement"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/fraud"
	"rekberkuy/core-service/internal/kyc"
	"rekberkuy/core-service/internal/midtrans"
	"rekberkuy/core-service/internal/relayer"
	"rekberkuy/core-service/internal/repository"
	"rekberkuy/core-service/internal/usecase"
	"rekberkuy/core-service/internal/worker"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to create database connection via GORM: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get sql.DB instance from GORM: %v", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeHrs) * time.Hour)

	fmt.Println("🚀 Go GORM Backend successfully connected to Supabase database!")

	// ============================================================================
	// DATABASE SCHEMA — managed by SQL migrations (golang-migrate)
	// ----------------------------------------------------------------------------
	// AutoMigrate has been retired in favour of manual versioned SQL migrations.
	// Apply the schema before running the server:
	//   go run ./cmd/migrate/main.go up
	// See db/migrations/ for the DDL and cmd/migrate/main.go for the runner.
	// ============================================================================

	// ============================================================================
	// DEPENDENCY INJECTION (REPOS, USECASES, HANDLERS)
	// ============================================================================

	// 1. Repository Layer
	walletRepo := repository.NewWalletRepository(sqlDB)
	transactionRepo := repository.NewTransactionRepository(sqlDB)
	userRepo := repository.NewUserRepository(sqlDB)
	kycRepo := repository.NewKYCRepository(sqlDB)
	vendorRepo := repository.NewVendorRepository(sqlDB)
	categoryRepo := repository.NewCategoryRepository(sqlDB)
	reviewRepo := repository.NewReviewRepository(sqlDB)
	disputeRepo := repository.NewDisputeRepository(sqlDB)

	// Unit of Work: transactional boundary across repositories (Serializable + FOR UPDATE)
	unitOfWork := repository.NewUnitOfWork(sqlDB)

	// External adapters: fraud scoring + on-chain audit-log relayer (gasless) + Midtrans.
	// Falls back to stub when an external service is not configured (dev/test env).
	fraudClient := newFraudClient(cfg)
	kycClient := newKYCClient(cfg)
	relayerSvc := newRelayer(cfg)
	midtransClient := newMidtransClient(cfg)

	// Idempotency: prefer Redis when configured, fall back to PostgreSQL.
	idemRepo := setupIdempotency(context.Background(), cfg, sqlDB)
	// 2. Usecase Layer
	financeCalc := usecase.NewFinanceCalculator()
	userUsecase := usecase.NewUserUsecase(unitOfWork, userRepo, walletRepo, midtransClient)
	kycUsecase := usecase.NewKYCUsecase(kycRepo, unitOfWork, kycClient)
	vendorUsecase := usecase.NewVendorUsecase(vendorRepo)
	reviewUsecase := usecase.NewReviewUsecase(transactionRepo, reviewRepo)
	disbursementUsecase := usecase.NewDisbursementUsecase(transactionRepo)
	disputeUsecase := usecase.NewDisputeUsecase(unitOfWork, disputeRepo)

	// Auth: token signing service + credential usecase (register/login).
	tokenLifetime, err := time.ParseDuration(cfg.JWT.TokenLifetime)
	if err != nil {
		log.Printf("⚠️  Invalid JWT_TOKEN_LIFETIME %q, defaulting to 24h", cfg.JWT.TokenLifetime)
		tokenLifetime = 24 * time.Hour
	}
	tokenService := usecase.NewTokenService(cfg.JWT.Secret, tokenLifetime)
	authUsecase := usecase.NewAuthUsecase(unitOfWork, userRepo, tokenService)

	goodsUsecase := usecase.NewTransactionGoodsUsecase(unitOfWork, transactionRepo, financeCalc, fraudClient, relayerSvc)
	servicesUsecase := usecase.NewTransactionServicesUsecase(unitOfWork, transactionRepo, financeCalc, fraudClient, relayerSvc)
	eventsUsecase := usecase.NewTransactionEventsUsecase(unitOfWork, transactionRepo, walletRepo, financeCalc, fraudClient, relayerSvc)

	// Read APIs + withdrawal flow. The disbursement client is the seam for the
	// future Midtrans Payout integration; today it serves the ESTIMATE half of
	// the withdrawal true-up model (stub = configured flat fee).
	disbursementClient := disbursement.NewDisbursementStub(cfg.Midtrans.DisbursementFee)
	transactionQueryUsecase := usecase.NewTransactionQueryUsecase(transactionRepo)
	catalogUsecase := usecase.NewCatalogUsecase(categoryRepo, vendorRepo)
	withdrawalUsecase := usecase.NewWithdrawalUsecase(unitOfWork, walletRepo, disbursementClient)

	// 3. Handler Layer
	userHandler := handlers.NewUserHandler(userUsecase, cfg.JWT.Secret)
	authHandler := handlers.NewAuthHandler(authUsecase)
	walletHandler := handlers.NewWalletHandler(userUsecase, withdrawalUsecase)
	kycHandler := handlers.NewKYCHandler(kycUsecase)
	vendorHandler := handlers.NewVendorHandler(vendorUsecase)
	reviewHandler := handlers.NewReviewHandler(reviewUsecase)
	disbursementHandler := handlers.NewDisbursementHandler(disbursementUsecase)
	disputeHandler := handlers.NewDisputeHandler(disputeUsecase)
	transactionQueryHandler := handlers.NewTransactionQueryHandler(transactionQueryUsecase)
	catalogHandler := handlers.NewCatalogHandler(catalogUsecase)

	goodsHandler := handlers.NewTransactionGoodsHandler(goodsUsecase)
	servicesHandler := handlers.NewTransactionServicesHandler(servicesUsecase)
	eventsHandler := handlers.NewTransactionEventsHandler(eventsUsecase)

	webhookHandler := handlers.NewMidtransWebhookHandler(midtransClient, userUsecase, goodsUsecase, servicesUsecase, eventsUsecase)

	// Auth middleware (JWT secret from config, not os.Getenv per-request)
	authMW := handlers.NewAuthMiddleware(cfg.JWT.Secret)

	// ============================================================================
	// BACKGROUND WORKERS
	// ============================================================================
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()

	releaseWorker := worker.NewAutoReleaseWorker(transactionRepo, goodsUsecase)
	releaseWorker.Start(workerCtx)

	crmWorker := worker.NewCRMWorker(walletRepo, reviewRepo, unitOfWork, financeCalc)
	crmWorker.Start(workerCtx)

	// ============================================================================
	// HTTP ROUTING & MIDDLEWARE
	// ============================================================================
	r := gin.Default()
	r.Use(handlers.CORSMiddleware(cfg.App.CORSAllowedOrigins))
	r.Use(handlers.IdempotencyMiddleware(idemRepo))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Rekberkuy Engine is running smoothly with Multi-Tenant Architecture",
		})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		// Public auth routes (credential-based register/login)
		api.POST("/auth/register", authHandler.RegisterHandler)
		api.POST("/auth/login", authHandler.LoginHandler)

		// Dev-only token backdoor — never registered in production.
		if !cfg.App.IsProduction() {
			api.POST("/users/token-test", userHandler.GenerateTokenTestHandler)
		}

		// KYC & vendor (require login)
		api.POST("/kyc/submit", authMW.RequireRole(domain.RoleUser), kycHandler.SubmitKYCHandler)
		api.POST("/vendors/register", authMW.RequireRole(domain.RoleUser), vendorHandler.RegisterVendorHandler)

		// KYC admin review — the AI score stored on each submission is a
		// REFERENCE only; approve/reject is always the admin's decision.
		adminGroup := api.Group("/admin", authMW.RequireRole(domain.RoleAdmin))
		{
			adminGroup.GET("/kyc/pending", kycHandler.GetPendingKYCsHandler)
			adminGroup.GET("/kyc/:id", kycHandler.GetKYCDetailHandler)
			adminGroup.POST("/kyc/:id/review", kycHandler.ReviewKYCHandler)
			adminGroup.GET("/withdrawals/pending", walletHandler.ListPendingWithdrawalsHandler)
			adminGroup.POST("/withdrawals/:id/disburse", walletHandler.MarkWithdrawalDisbursedHandler)
		}

		// Read APIs: transaction list/detail (party-scoped)
		api.GET("/transactions", authMW.RequireRole(domain.RoleUser), transactionQueryHandler.ListMyTransactionsHandler)
		api.GET("/transactions/:id", authMW.RequireRole(domain.RoleUser), transactionQueryHandler.GetTransactionDetailHandler)

		// Public catalog: 3-tier taxonomy + vendor marketplace
		api.GET("/categories", catalogHandler.GetCategoryCatalogHandler)
		api.GET("/vendors", catalogHandler.ListMarketplaceVendorsHandler)

		// Goods Transactions
		goodsGroup := api.Group("/transactions/goods")
		{
			goodsGroup.POST("/lock", authMW.RequireRole(domain.RoleUser), goodsHandler.LockFundsGoodsHandler)
			goodsGroup.POST("/release", authMW.RequireRole(domain.RoleUser), goodsHandler.ReleaseGoodsHandler)
		}

		// Services Transactions
		servicesGroup := api.Group("/transactions/services")
		{
			servicesGroup.POST("/lock", authMW.RequireRole(domain.RoleUser), servicesHandler.LockFundsServicesHandler)
			servicesGroup.POST("/release-milestone", authMW.RequireRole(domain.RoleUser), servicesHandler.ReleaseMilestoneHandler)
		}

		// Event Transactions
		eventsGroup := api.Group("/transactions/events")
		{
			eventsGroup.POST("/lock", authMW.RequireRole(domain.RoleUser), eventsHandler.LockFundsEventsHandler)
			eventsGroup.POST("/:id/vendor-invoices", authMW.RequireRole(domain.RoleEventOrganizer, domain.RoleAdmin), eventsHandler.SubmitVendorInvoiceHandler)
			eventsGroup.POST("/release-milestone", authMW.RequireRole(domain.RoleAdmin), eventsHandler.ReleaseEventMilestoneHandler)
			eventsGroup.POST("/release-vendors", authMW.RequireRole(domain.RoleEventOrganizer, domain.RoleAdmin), eventsHandler.ProcessEventVendorPayoutHandler)
			eventsGroup.POST("/payouts/:id/disburse", authMW.RequireRole(domain.RoleAdmin), disbursementHandler.MarkDisbursedHandler)
		}

		// Wallet
		wallets := api.Group("/wallets")
		{
			wallets.POST("/topup", authMW.RequireRole(domain.RoleUser), walletHandler.CreateTopUpHandler)
			wallets.GET("/me", authMW.RequireRole(domain.RoleUser), walletHandler.GetBalanceHandler)
			wallets.GET("/me/transactions", authMW.RequireRole(domain.RoleUser), walletHandler.GetHistoryHandler)
			wallets.POST("/withdraw", authMW.RequireRole(domain.RoleUser), walletHandler.RequestWithdrawalHandler)
			wallets.GET("/withdrawals", authMW.RequireRole(domain.RoleUser), walletHandler.ListMyWithdrawalsHandler)
		}

		// Reviews (buyer rates counterparty after RELEASED)
		api.POST("/reviews", authMW.RequireRole(domain.RoleUser), reviewHandler.CreateReviewHandler)
		api.GET("/users/:id/reviews", reviewHandler.GetReviewsForUserHandler)

		// Disputes (raise while funds locked; Admin mediates — ADR-0003)
		api.POST("/disputes", authMW.RequireRole(domain.RoleUser), disputeHandler.OpenDisputeHandler)
		api.POST("/disputes/:id/acknowledge", authMW.RequireRole(domain.RoleAdmin), disputeHandler.AcknowledgeDisputeHandler)
		api.POST("/disputes/:id/resolve", authMW.RequireRole(domain.RoleAdmin), disputeHandler.ResolveDisputeHandler)
		api.GET("/disputes/:id", disputeHandler.GetDisputeHandler)

		// Midtrans webhook (public, no JWT — verified via SignatureKey)
		api.POST("/webhooks/midtrans", webhookHandler.NotificationHandler)
	}

	// ============================================================================
	// GRACEFUL SHUTDOWN
	// ============================================================================
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	// context cancelled when SIGINT/SIGTERM is received
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server running on port %s...", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server failed to run: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, stopping server safely...")

	// Stop background workers
	stopWorkers()
	releaseWorker.Stop()
	crmWorker.Stop()

	// Give in-flight requests time to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️  Server shutdown encountered an error: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("⚠️  Database connection close encountered an error: %v", err)
	}

	log.Println("Server successfully shut down. Goodbye! 👋")
}

// setupIdempotency selects the idempotency backend: Redis when configured &
// reachable, otherwise PostgreSQL (race-safe via ON CONFLICT).
func setupIdempotency(ctx context.Context, cfg *config.Config, sqlDB *sql.DB) domain.IdempotencyRepository {
	if cfg.RedisEnabled() {
		client, err := repository.NewRedisClient(ctx, cfg.Redis.URL)
		if err != nil {
			log.Printf("⚠️  Redis unreachable (%v), falling back to PostgreSQL idempotency.", err)
		} else {
			log.Println("🔌 Idempotency backend: Redis (SETNX + TTL)")
			return repository.NewIdempotencyRedisRepository(client, cfg.Redis.TTL.IdempotencySec)
		}
	}
	log.Println("🔌 Idempotency backend: PostgreSQL (ON CONFLICT)")
	return repository.NewIdempotencyRepository(sqlDB)
}

// newFraudClient selects the fraud scoring adapter: HTTP to the backend-ai
// verification service when screening is enabled, else the always-safe stub.
//
// backend-ai only RETURNS A SCORE; the adapter (this process) makes the isSafe
// decision using cfg.AI.UnsafeThreshold — the backend is the decision maker.
func newFraudClient(cfg *config.Config) domain.FraudClient {
	if cfg.AI.ScreeningEnabled {
		log.Printf("🔌 Fraud backend: backend-ai (%s), decision threshold=%.2f", cfg.AI.ServiceURL, cfg.AI.UnsafeThreshold)
		return fraud.NewFraudHTTPClient(cfg.AI.ServiceURL, 5*time.Second, cfg.AI.FailOpen, cfg.AI.UnsafeThreshold)
	}
	log.Println("🔌 Fraud backend: stub (FRAUD_SCREENING_ENABLED=false)")
	return fraud.NewFraudClientStub()
}

// newKYCClient selects the KYC verification adapter: HTTP to the backend-ai
// service when screening is enabled, else the always-unavailable stub. Unlike
// fraud, an unavailable KYC AI is NOT fatal — the submission simply proceeds
// without an AI reference and the admin reviews the raw documents.
func newKYCClient(cfg *config.Config) domain.KYCClient {
	if cfg.AI.ScreeningEnabled {
		log.Printf("🔌 KYC AI backend: backend-ai (%s), reference-only scoring", cfg.AI.ServiceURL)
		return kyc.NewKYCHTTPClient(cfg.AI.ServiceURL, 10*time.Second)
	}
	log.Println("🔌 KYC AI backend: stub (FRAUD_SCREENING_ENABLED=false)")
	return kyc.NewKYCClientStub()
}

// newRelayer selects the on-chain audit-log adapter: go-ethereum relayer to Avalanche
// when fully configured, falls back to stub (dummy tx hash) for development.
func newRelayer(cfg *config.Config) domain.Relayer {
	if cfg.BlockchainEnabled() {
		r, err := relayer.NewEthRelayer(
			cfg.Blockchain.AvalancheRPCURL,
			cfg.Blockchain.DeployerPrivateKey,
			cfg.Blockchain.ContractAddress,
			cfg.Blockchain.ChainID,
		)
		if err != nil {
			log.Printf("⚠️  Failed to init on-chain relayer (%v), falling back to stub.", err)
			return relayer.NewRelayerStub()
		}
		log.Println("⛓️  Relayer on-chain: Avalanche (gasless audit-log)")
		return r
	}
	log.Println("⛓️  Relayer on-chain: stub (blockchain not configured)")
	return relayer.NewRelayerStub()
}

// newMidtransClient selects the Midtrans adapter: real Snap client when ServerKey
// is configured, nil otherwise (top-up will be rejected with a clear message).
func newMidtransClient(cfg *config.Config) domain.MidtransClient {
	if cfg.MidtransEnabled() {
		log.Printf("💳 Midtrans: %s", cfg.Midtrans.Environment)
		return midtrans.NewSnapClient(cfg.Midtrans.ServerKey, cfg.Midtrans.Environment, 10*time.Second)
	}
	log.Println("💳 Midtrans: not configured (top-up inactive)")
	return nil
}
