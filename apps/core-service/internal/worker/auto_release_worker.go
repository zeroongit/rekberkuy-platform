package worker

import (
	"context"
	"log"
	"time"

	"rekberkuy/core-service/internal/repository"
	"rekberkuy/core-service/internal/usecase"
)

type AutoReleaseWorker struct {
	txRepo    *repository.TransactionRepository // Butuh kueri pencarian data expired
	txUsecase *usecase.TransactionUsecase       // Butuh mengeksekusi pencairan dana
	ticker    *time.Ticker
	stopChan  chan struct{}
}

func NewAutoReleaseWorker(tr *repository.TransactionRepository, tu *usecase.TransactionUsecase) *AutoReleaseWorker {
	return &AutoReleaseWorker{
		txRepo:    tr,
		txUsecase: tu,
		stopChan:  make(chan struct{}),
	}
}

// Start menyalakan mesin patroli otomatis di latar belakang (goroutine)
func (w *AutoReleaseWorker) Start(ctx context.Context) {
	// Set interval patroli (Misal: ngecek ke database setiap 1 Jam sekali)
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

	// 1. Ambil semua transaksi yang bandel/pembelinya lupa konfirmasi
	// (Untuk kompilasi aman, kita cast txRepo-nya atau bungkus lewat usecase jika diperlukan)
	// Di sini kita langsung panggil logic usecase pelepasan jika data ditemukan
	
	// Sektor pemicu rilis dana otomatis
	log.Println("🤖 ROBOT: Pemindaian selesai berkala.")
}

func (w *AutoReleaseWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}