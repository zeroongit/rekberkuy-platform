package worker

import (
	"context"
	"log"
	"time"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

type AutoReleaseWorker struct {
	txRepo       domain.TransactionRepository
	goodsUsecase *usecase.TransactionGoodsUsecase
	ticker       *time.Ticker
	stopChan     chan struct{}
}

// NewAutoReleaseWorker initializes the scanner bot for expired goods transactions
func NewAutoReleaseWorker(tr domain.TransactionRepository, gu *usecase.TransactionGoodsUsecase) *AutoReleaseWorker {
	return &AutoReleaseWorker{
		txRepo:       tr,
		goodsUsecase: gu,
		stopChan:     make(chan struct{}),
	}
}

func (w *AutoReleaseWorker) Start(ctx context.Context) {
	w.ticker = time.NewTicker(1 * time.Hour)
	log.Println("🤖 ROBOT: RekberKuy Auto-Release Engine successfully started...")

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.executeAutoRelease(ctx)
			case <-w.stopChan:
				log.Println("🤖 ROBOT: Auto-Release Engine safely stopped.")
				return
			}
		}
	}()
}

func (w *AutoReleaseWorker) executeAutoRelease(ctx context.Context) {
	log.Println("🤖 ROBOT: Starting scan for goods transactions past the confirmation deadline...")

	ids, err := w.txRepo.GetExpiredLockedTransactions(ctx)
	if err != nil {
		log.Printf("🤖 ROBOT ERROR: Failed to scan expired data from repository: %v", err)
		return
	}

	if len(ids) == 0 {
		log.Println("🤖 ROBOT: No expired transactions pending this hour.")
		return
	}

	log.Printf("🤖 ROBOT: Found %d expired goods transactions ready for auto-release.", len(ids))

	for _, txID := range ids {
		log.Printf("🤖 ROBOT: Forcing fund release for Transaction ID: %s", txID)
		err := w.goodsUsecase.ReleaseFundsGoods(ctx, txID)
		if err != nil {
			log.Printf("🤖 ROBOT ERROR: Failed to auto-release funds for transaction %s: %v", txID, err)
			continue
		}
		log.Printf("🤖 ROBOT SUCCESS: Funds for transaction %s successfully auto-released to seller.", txID)
	}

	log.Println("🤖 ROBOT: All periodic scan cycles completed.")
}

func (w *AutoReleaseWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}
