package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rekberkuy/core-backend/internal/domain"
)

type transactionUsecase struct {
	walletRepo  domain.WalletRepository
	financeCalc *FinanceCalculator
}

// NewTransactionUsecase menginisialisasi manajer pengatur alur transaksi Rekberkuy
func NewTransactionUsecase(wr domain.WalletRepository, fc *FinanceCalculator) *transactionUsecase {
	return &transactionUsecase{
		walletRepo:  wr,
		financeCalc: fc,
	}
}

// LockFundsAwal menangani alur ketika pembeli/peserta mengunci dana mereka ke escrow Rekberkuy
func (u *transactionUsecase) LockFundsAwal(ctx context.Context, buyerID string, amountBase int64, rekberType domain.RekberType, isRekberPay bool, sellerTier string) (*domain.RekberPayTransaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("nominal transaksi harus lebih besar dari nol")
	}

	// 1. Panggil kalkulator pintar untuk menghitung Buyer Service Fee berdasarkan kasta seller
	buyerFee := u.financeCalc.CalculateBuyerServiceFee(rekberType, amountBase, isRekberPay, sellerTier)
	totalDipotong := amountBase + buyerFee

	// 2. Tampung description ke variabel agar bisa diambil pointer-nya (&descMsg)
	descMsg := fmt.Sprintf("Penguncian dana escrow untuk transaksi jenis %s", rekberType)

	// 3. Siapkan objek mutasi log dengan melakukan Type Casting ke domain.WalletTxStatus
	txRecord := &domain.RekberPayTransaction{
		ID:          fmt.Sprintf("TX-LOCK-%d", time.Now().UnixNano()),
		WalletID:    buyerID,
		Type:        "PAYMENT_ESCROW",
		Status:      domain.WalletTxStatus(domain.StatusFundsLocked), // FIX: Di-cast ke WalletTxStatus agar sinkron
		Amount:      totalDipotong,
		AdminFee:    buyerFee,
		Description: &descMsg,
	}

	// 4. Perintahkan WalletRepository untuk mengeksekusi mutasi aman (ACID + Anti Race Condition)
	err := u.walletRepo.UpdateBalanceTx(ctx, txRecord, -totalDipotong)
	if err != nil {
		return nil, fmt.Errorf("gagal mengunci dana di escrow: %w", err)
	}

	return txRecord, nil
}

// ReleaseFundsEventSelesai menangani proses audit akhir khusus Event ketika acara selesai.
func (u *transactionUsecase) ReleaseFundsEventSelesai(ctx context.Context, totalEscrowLocked int64, vendorsBill []domain.EventVendorAllocation, eoID string, eoTier string) (*domain.EventAuditResult, error) {
	
	// 1. Panggil kalkulator pintar untuk melakukan audit pemecahan dana secara adil sesuai aturan main kita
	auditResult := u.financeCalc.CalculateEventAudit(totalEscrowLocked, vendorsBill, eoTier)

	// 2. Eksekusi Pencairan Dana ke EO jika mereka berhak mendapatkan bonus efisiensi (BonusToEO > 0)
	if auditResult.BonusToEO > 0 {
		descBonus := fmt.Sprintf("Bonus legal efisiensi anggaran event sebesar %s persen", eoTier)

		eoTxLog := &domain.RekberPayTransaction{
			ID:                fmt.Sprintf("TX-BONUS-EO-%d", time.Now().UnixNano()),
			WalletID:          eoID,
			Type:              "EO_EFFICIENCY_BONUS",
			Status:            domain.WalletTxStatus(domain.StatusReleased), // FIX: Di-cast ke WalletTxStatus agar sinkron
			Amount:            auditResult.BonusToEO,
			PlatformNetProfit: auditResult.PlatformFee,
			Description:       &descBonus,
		}
		
		// Kirim nilai POSITIF karena saldo wallet RekberPay milik EO bertambah
		err := u.walletRepo.UpdateBalanceTx(ctx, eoTxLog, auditResult.BonusToEO)
		if err != nil {
			return nil, fmt.Errorf("gagal mencairkan bonus efisiensi ke dompet EO: %w", err)
		}
	}

	return &auditResult, nil
}