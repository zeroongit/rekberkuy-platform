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

// NewAutoReleaseWorker menginisialisasi robot pemindai transaksi barang expired
func NewAutoReleaseWorker(tr domain.TransactionRepository, gu *usecase.TransactionGoodsUsecase) *AutoReleaseWorker {
	return &AutoReleaseWorker{
		txRepo:       tr,
		goodsUsecase: gu,
		stopChan:     make(chan struct{}),
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
	log.Println("🤖 ROBOT: Memulai pemindaian transaksi barang yang melewati batas konfirmasi...")
	
	ids, err := w.txRepo.GetExpiredLockedTransactions(ctx)
	if err != nil {
		log.Printf("🤖 ROBOT ERROR: Gagal memindai data expired dari repositori: %v", err)
		return
	}

	if len(ids) == 0 {
		log.Println("🤖 ROBOT: Tidak ada transaksi expired yang menggantung jam ini.")
		return
	}

	log.Printf("🤖 ROBOT: Menemukan %d transaksi barang expired siap di-release otomatis.", len(ids))

	for _, txID := range ids {
		log.Printf("🤖 ROBOT: Menjalankan pemaksaan release dana untuk Transaksi ID: %s", txID)
		err := w.goodsUsecase.ReleaseFundsGoods(ctx, txID) 
		if err != nil {
			log.Printf("🤖 ROBOT ERROR: Gagal melepas dana otomatis untuk transaksi %s: %v", txID, err)
			continue
		}
		log.Printf("🤖 ROBOT SUCCESS: Dana transaksi %s berhasil dilepas otomatis ke penjual.", txID)
	}

	log.Println("🤖 ROBOT: Seluruh siklus pemindaian berkala selesai.")
}

func (w *AutoReleaseWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}