package worker

import (
	"context"
	"log"
	"time"

	"rekberkuy/core-service/internal/domain" // IMPORT INTERFACE DOMAIN
	"rekberkuy/core-service/internal/usecase"
)

type AutoReleaseWorker struct {
	txRepo    domain.TransactionRepository // UBAH JADI INTERFACE
	txUsecase *usecase.TransactionUsecase
	ticker    *time.Ticker
	stopChan  chan struct{}
}

// Ganti parameter tipe pertama menjadi domain.TransactionRepository
func NewAutoReleaseWorker(tr domain.TransactionRepository, tu *usecase.TransactionUsecase) *AutoReleaseWorker {
	return &AutoReleaseWorker{
		txRepo:    tr,
		txUsecase: tu,
		stopChan:  make(chan struct{}),
	}
}

func (w *AutoReleaseWorker) Start(ctx context.Context) {
	w.ticker = time.NewTicker(1 * time.Hour)
	log.Println("🤖 ROBOT: Auto-Release Engine RekberKuy berhasil dinyalakan...")

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.executeAutoRelease(ctx)
			case <-w.stopChan:
				log.Println("🤖 ROBOT: Auto-Release Engine dimatikan secara aman.")
				return
			}
		}
	}()
}

func (w *AutoReleaseWorker) executeAutoRelease(ctx context.Context) {
	log.Println("🤖 ROBOT: Memulai pemindaian transaksi hantu yang melewati batas 3 hari...")
	
	ids, err := w.txRepo.GetExpiredLockedTransactions(ctx)
	if err != nil {
		log.Printf("🤖 ROBOT ERROR: Gagal memindai data expired: %v", err)
		return
	}

	log.Printf("🤖 ROBOT: Menemukan %d transaksi expired siap di-release otomatis.", len(ids))
	log.Println("🤖 ROBOT: Pemindaian selesai berkala.")
}

func (w *AutoReleaseWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}