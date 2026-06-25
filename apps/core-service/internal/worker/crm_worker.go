package worker

import (
	"context"
	"log"
	"time"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type crmWorker struct {
	walletRepo  domain.WalletRepository
	calculator  *usecase.FinanceCalculator
	stopChannel chan struct{}
}

// NewCRMWorker bertindak sebagai constructor untuk menginisialisasi Worker (Patuhi kontrak domain.CRMWorker)
func NewCRMWorker(walletRepo domain.WalletRepository, calculator *usecase.FinanceCalculator) domain.CRMWorker {
	return &crmWorker{
		walletRepo:  walletRepo,
		calculator:  calculator,
		stopChannel: make(chan struct{}),
	}
}

func (w *crmWorker) Start(ctx context.Context) {
	log.Println("[WORKER] CRM Loyalty Evaluation Worker Engine has been successfully initialized.")

	go func() {
		tickerDuration := w.calculateDurationToNextMonthFirstDay()
		timer := time.NewTimer(tickerDuration)
		defer timer.Stop()

		log.Printf("[WORKER] First evaluation scheduled in %v", tickerDuration)

		for {
			select {
			case <-timer.C:
				evalCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if err := w.ExecuteMonthlyEvaluation(evalCtx); err != nil {
					log.Printf("[WORKER_ERROR] Failed executing monthly CRM evaluation: %v", err)
				}
				cancel()

				timer.Reset(w.calculateDurationToNextMonthFirstDay())

			case <-w.stopChannel:
				log.Println("[WORKER] Worker engine received shutdown signal. Stopping safely...")
				return
			case <-ctx.Done():
				log.Println("[WORKER] Global context canceled. Stopping worker...")
				return
			}
		}
	}()
}

func (w *crmWorker) Stop() {
	close(w.stopChannel)
}

func (w *crmWorker) ExecuteMonthlyEvaluation(ctx context.Context) error {
	log.Println("[WORKER] Initiating monthly CRM evaluation batch transaction...")

	// 1. Ambil pointer data user-profile murni (*[]*domain.UserProfile)
	users, err := w.walletRepo.GetAllUsersForCRMEvaluation(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		// Menggunakan urutan transaksi terisolasi ACID per user
		err := w.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
			// 2. Cari data status CRM target
			crmProfile, err := txRepo.GetCRMLoyaltyByUserID(ctx, user.ID)
			if err != nil {
				log.Printf("[WORKER_SKIP] User ID %s doesn't have active CRM table log. Skipping.", user.ID)
				return nil
			}

			// 3. Logika multi-role adaptif ala shopee
			rekberType := domain.TypeGoods
			if user.Role == domain.RoleEventOrganizer {
				rekberType = domain.TypeEvents
			} else if user.Role == domain.RoleVerifiedVendor {
				rekberType = domain.TypeServices
			}

			// Rating aman default anti-fraud
			currentRating := 5.0

			// 4. Hitung kasta menggunakan calculator yang terpisah modular
			newTier, statusMsg := w.calculator.EvaluateMonthlyMerchantTier(rekberType, *crmProfile, currentRating)

			oldTier := crmProfile.CurrentTier
			crmProfile.CurrentTier = newTier

			// Akumulasi sanksi low sales jika performa turun
			if statusMsg == "WARNING_LOW_SALES" {
				crmProfile.ConsecutiveFailedMonths++
			} else if statusMsg == "STAY_GOLD" || statusMsg == "UPGRADE_TO_GOLD" || statusMsg == "UPGRADE_TO_SILVER" {
				crmProfile.ConsecutiveFailedMonths = 0 // Reset jika penjualan pulih
			}

			// 5. Simpan kembali kasta BRONZE/SILVER/GOLD yang valid ke database
			err = txRepo.UpdateCRMLoyalty(ctx, crmProfile)
			if err != nil {
				return err
			}

			log.Printf("[WORKER_SUCCESS] Processed User %s (%s). Action: %s | Old Tier: %s -> New Tier: %s",
				user.ID, user.Role, statusMsg, oldTier, crmProfile.CurrentTier)
			return nil
		})

		if err != nil {
			log.Printf("[WORKER_ERROR] Failed processing loyalty updates for user %s: %v", user.ID, err)
			continue
		}
	}

	log.Println("[WORKER] Monthly CRM evaluation execution batch has finished successfully.")
	return nil
}

func (w *crmWorker) calculateDurationToNextMonthFirstDay() time.Duration {
	now := time.Now()
	nextMonth := now.AddDate(0, 1, -now.Day()+1)
	nextMonthFirstDay := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, now.Location())
	return nextMonthFirstDay.Sub(now)
}