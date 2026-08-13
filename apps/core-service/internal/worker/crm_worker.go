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
	reviewRepo  domain.ReviewRepository
	uow         domain.UnitOfWork
	calculator  *usecase.FinanceCalculator
	stopChannel chan struct{}
}

// NewCRMWorker acts as a constructor to initialize the Worker (comply with the domain.CRMWorker contract)
func NewCRMWorker(walletRepo domain.WalletRepository, reviewRepo domain.ReviewRepository, uow domain.UnitOfWork, calculator *usecase.FinanceCalculator) domain.CRMWorker {
	return &crmWorker{
		walletRepo:  walletRepo,
		reviewRepo:  reviewRepo,
		uow:         uow,
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

	// 1. Fetch all user profiles (read, outside the transactional boundary)
	users, err := w.walletRepo.GetAllUsersForCRMEvaluation(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		// Evaluate each user in its own isolated ACID transaction unit
		err := w.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
			crmProfile, err := stores.Wallets.GetCRMLoyaltyByUserID(ctx, user.ID)
			if err != nil {
				log.Printf("[WORKER_SKIP] User ID %s doesn't have active CRM table log. Skipping.", user.ID)
				return nil
			}

			// Adaptive multi-role logic
			rekberType := domain.TypeGoods
			if user.Role == domain.RoleEventOrganizer {
				rekberType = domain.TypeEvents
			} else if user.Role == domain.RoleVerifiedVendor {
				rekberType = domain.TypeServices
			}

			// Real average rating from reviews; fall back to a neutral 5.0 when the
			// merchant has no reviews yet (or the read fails).
			currentRating := 5.0
			if w.reviewRepo != nil {
				if avg, err := w.reviewRepo.GetAverageRatingForUser(ctx, user.ID); err == nil && avg > 0 {
					currentRating = avg
				}
			}

			newTier, statusMsg := w.calculator.EvaluateMonthlyMerchantTier(rekberType, *crmProfile, currentRating)

			oldTier := crmProfile.CurrentTier
			crmProfile.CurrentTier = newTier

			if statusMsg == "WARNING_LOW_SALES" {
				crmProfile.ConsecutiveFailedMonths++
			} else if statusMsg == "STAY_GOLD" || statusMsg == "UPGRADE_TO_GOLD" || statusMsg == "UPGRADE_TO_SILVER" {
				crmProfile.ConsecutiveFailedMonths = 0
			}

			if err := stores.Wallets.UpdateCRMLoyalty(ctx, crmProfile); err != nil {
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
