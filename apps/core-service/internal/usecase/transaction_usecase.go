package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid" 
	"rekberkuy/core-service/internal/domain"
)

type TransactionUsecase struct {
	transactionRepo domain.TransactionRepository 
	walletRepo      domain.WalletRepository
	financeRepo     domain.FinanceRepository 
	financeCalc     *FinanceCalculator
}

// NewTransactionUsecase menginisialisasi manajer pengatur alur transaksi Rekberkuy
func NewTransactionUsecase(tr domain.TransactionRepository, wr domain.WalletRepository, fr domain.FinanceRepository, fc *FinanceCalculator) *TransactionUsecase {
	return &TransactionUsecase{
		transactionRepo: tr,
		walletRepo:      wr,
		financeRepo:     fr,
		financeCalc:     fc,
	}
}

// LockFundsAwal menangani alur ketika pembeli mengunci dana mereka ke escrow Rekberkuy
func (u *TransactionUsecase) LockFundsAwal(ctx context.Context, buyerID string, sellerID string, amountBase int64, rekberType domain.RekberType, isRekberPay bool, sellerTier string, shippingFee int64, paymentMethod string, idempotencyKey string) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("nominal transaksi harus lebih besar dari nol")
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(rekberType, amountBase, isRekberPay, sellerTier)
	sellerFee := u.financeCalc.CalculateSellerServiceFee(rekberType, amountBase, sellerTier)

	amountGross := amountBase + buyerFee + shippingFee
	amountNet := amountBase - sellerFee

	midtransOrderID := fmt.Sprintf("REKBERKUY-ORDER-%s", uuid.New().String()[:8])

	txMaster := &domain.Transaction{
		ID:               uuid.New().String(),
		BuyerID:          buyerID,
		SellerID:         sellerID,
		Type:             rekberType,
		Status:           domain.StatusWaitingPayment, 
		AmountBase:       amountBase,
		ShippingFee:      shippingFee,
		ServiceFee:       buyerFee,
		MidtransFee:      0, // Nanti diisi riil saat webhook settlement memotong
		AmountGross:      amountGross,
		AmountNet:        amountNet,
		MidtransOrderID:  midtransOrderID,
		IdempotencyKey:   idempotencyKey,
		PaymentMethod:    paymentMethod,
		BlockchainTxHash: nil, 
	}
	
	err := u.transactionRepo.CreateTransaction(ctx, txMaster)
	if err != nil {
		return nil, fmt.Errorf("gagal mencatat transaksi master escrow: %w", err)
	}

	return txMaster, nil
}

// ConfirmPaymentWebhookMidtrans memproses perpindahan State Machine dari WAITING_PAYMENT ke FUNDS_LOCKED
func (u *TransactionUsecase) ConfirmPaymentWebhookMidtrans(ctx context.Context, transactionID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}

		if tx.Status != domain.StatusWaitingPayment {
			return fmt.Errorf("transaksi tidak dapat diproses: status saat ini adalah %s", tx.Status)
		}

		descMsg := fmt.Sprintf("Pembayaran sukses untuk transaksi escrow #%s", tx.ID)
		walletTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.BuyerID,
			Type:        domain.TxPayment, 
			Status:      domain.WalletStatusSuccess,
			Amount:      tx.AmountGross,
			AdminFee:    tx.ServiceFee,
			Description: &descMsg,
		}

		err = txRepo.UpdateBalanceTx(ctx, walletTxLog, -tx.AmountGross)
		if err != nil {
			return fmt.Errorf("gagal mendebet saldo pembeli: %w", err)
		}

		err = u.transactionRepo.UpdateTransactionStatus(ctx, tx.ID, domain.StatusFundsLocked)
		if err != nil {
			return fmt.Errorf("gagal merubah state transaksi menjadi FUNDS_LOCKED: %w", err)
		}

		// 📈 FINTECH AUDIT: Catat masuknya uang pembeli ke dalam penampungan global platform
		err = u.financeRepo.UpdatePlatformFinance(ctx, tx.AmountGross, 0, 0)
		if err != nil {
			return fmt.Errorf("gagal memperbarui laporan keuangan platform: %w", err)
		}

		return nil
	})
}

// ReleaseFundsSelesai memproses pelepasan dana escrow dari platform ke dompet milik penjual (Barang / Jasa Lengkap)
func (u *TransactionUsecase) ReleaseFundsSelesai(ctx context.Context, transactionID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}

		if tx.Status != domain.StatusFundsLocked {
			return fmt.Errorf("dana gagal dilepas: status transaksi wajib FUNDS_LOCKED, status saat ini %s", tx.Status)
		}

		descMsg := fmt.Sprintf("Penerimaan dana dari penyelesaian transaksi escrow #%s", tx.ID)
		sellerTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.SellerID,
			Type:        domain.TxReceiveFunds, 
			Status:      domain.WalletStatusSuccess,
			Amount:      tx.AmountNet,
			AdminFee:    0,
			Description: &descMsg,
		}

		err = txRepo.UpdateBalanceTx(ctx, sellerTxLog, tx.AmountNet)
		if err != nil {
			return fmt.Errorf("gagal mengredit saldo ke dompet penjual: %w", err)
		}

		err = u.transactionRepo.UpdateTransactionStatus(ctx, tx.ID, domain.StatusReleased)
		if err != nil {
			return fmt.Errorf("gagal merubah state transaksi menjadi RELEASED: %w", err)
		}

		// 📈 FINTECH AUDIT: Keluarkan dana escrow pembeli, lalu kunci pendapatan keuntungan (ServiceFee) ke dalam kas bersih platform
		escrowOutbound := -tx.AmountGross
		platformRevenue := tx.ServiceFee
		err = u.financeRepo.UpdatePlatformFinance(ctx, escrowOutbound, platformRevenue, tx.MidtransFee)
		if err != nil {
			return fmt.Errorf("gagal memutasi profit finansial platform: %w", err)
		}

		return nil
	})
}

// ReleaseMilestoneFunds memproses pelepasan dana escrow per termin/milestone khusus untuk transaksi JASA
func (u *TransactionUsecase) ReleaseMilestoneFunds(ctx context.Context, milestoneID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		
		// 1. Ambil data milestone riil dari repositori transaksi, bukan dummy data lagi!
		milestone, err := u.transactionRepo.GetMilestoneByID(ctx, milestoneID)
		if err != nil {
			return fmt.Errorf("gagal mengambil data milestone: %w", err)
		}

		if milestone.Status == "RELEASED" {
			return fmt.Errorf("transaksi ditolak: dana termin milestone ini sudah dicairkan sebelumnya")
		}

		// 2. Ambil transaksi induk untuk mengidentifikasi ID penjual asli
		tx, err := u.transactionRepo.GetTransactionByID(ctx, milestone.TransactionID)
		if err != nil {
			return fmt.Errorf("transaksi induk tidak ditemukan: %w", err)
		}

		descMsg := fmt.Sprintf("Pencairan dana milestone termin Jasa [%s] untuk Transaksi #%s", milestone.Title, tx.ID)
		sellerTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.SellerID, // Dinamis mengambil target ID penjual dari transaksi
			Type:        domain.TxReceiveFunds,
			Status:      domain.WalletStatusSuccess,
			Amount:      milestone.Amount,
			AdminFee:    0,
			Description: &descMsg,
		}

		// 3. Kreditkan dana termin ke dompet RekberPay milik freelancer
		err = txRepo.UpdateBalanceTx(ctx, sellerTxLog, milestone.Amount)
		if err != nil {
			return fmt.Errorf("gagal mencairkan dana termin milestone ke freelancer: %w", err)
		}

		// 4. Perbarui status baris milestone menjadi RELEASED di database
		err = u.transactionRepo.UpdateMilestoneStatus(ctx, milestoneID, "RELEASED")
		if err != nil {
			return fmt.Errorf("gagal memperbarui status termin database: %w", err)
		}

		// 📈 FINTECH AUDIT: Kurangi total dana mengendap (Escrow) global sebesar jatah termin yang keluar
		err = u.financeRepo.UpdatePlatformFinance(ctx, -milestone.Amount, 0, 0)
		if err != nil {
			return fmt.Errorf("gagal memutasi kas escrow platform: %w", err)
		}

		return nil
	})
}

// ProcessEventVendorPayouts mengeksekusi pembagian dana escrow akhir ke seluruh vendor acara secara atomik
func (u *TransactionUsecase) ProcessEventVendorPayouts(ctx context.Context, transactionID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}

		if tx.Status != domain.StatusFundsLocked {
			return fmt.Errorf("dana event gagal dipecah: status transaksi wajib FUNDS_LOCKED, status saat ini %s", tx.Status)
		}

		// 1. Ambil daftar tagihan seluruh sub-vendor yang terikat murni via repositori transaksi riil
		payouts, err := u.transactionRepo.GetEventVendorPayoutsByTxID(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("gagal mengambil data tagihan vendor: %w", err)
		}

		if len(payouts) == 0 {
			return fmt.Errorf("tidak ada data tagihan invoice vendor yang ditemukan untuk event ini")
		}

		var totalDisbursed int64 = 0

		// 2. Loop dan distribusikan kredit saldo ke sub-vendor lapangan secara asinkron/atomik
		for _, payout := range payouts {
			if payout.Status == "APPROVED" {
				continue // Lewati jika sudah pernah dicairkan di milestone parsial sebelumnya
			}

			descMsg := fmt.Sprintf("Pencairan dana penuh Event untuk Vendor: %s. Kebutuhan: %s", payout.VendorName, payout.ExpenseDescription)
			
			// Perhatian: Di sistem modern, WalletID diarahkan ke ID entitas akun sub-vendor terkait
			vendorTxLog := &domain.RekberPayTransaction{
				ID:          uuid.New().String(),
				WalletID:    payout.TransactionID, // Atau UUID spesifik akun dompet vendor lapangan Anda
				Type:        domain.TxReceiveFunds,
				Status:      domain.WalletStatusSuccess,
				Amount:      payout.AmountRequested,
				AdminFee:    0,
				Description: &descMsg,
			}

			err := txRepo.UpdateBalanceTx(ctx, vendorTxLog, payout.AmountRequested)
			if err != nil {
				return fmt.Errorf("gagal mentransfer dana ke vendor %s: %w", payout.VendorName, err)
			}

			// Perbarui status invoice pengajuan vendor menjadi APPROVED
			err = u.transactionRepo.UpdateEventVendorPayoutStatus(ctx, payout.ID, "APPROVED")
			if err != nil {
				return fmt.Errorf("gagal merubah status klaim vendor %s: %w", payout.VendorName, err)
			}

			totalDisbursed += payout.AmountRequested
		}

		// 3. Update status induk transaksi event menjadi RELEASED
		err = u.transactionRepo.UpdateTransactionStatus(ctx, transactionID, domain.StatusReleased)
		if err != nil {
			return fmt.Errorf("gagal merubah state transaksi induk event: %w", err)
		}

		// 📈 FINTECH AUDIT: Bersihkan sisa kas mengendap, kunci revenue keuntungan service fee EO/Platform
		err = u.financeRepo.UpdatePlatformFinance(ctx, -tx.AmountGross, tx.ServiceFee, tx.MidtransFee)
		if err != nil {
			return fmt.Errorf("gagal memutasi tutup buku kas event platform: %w", err)
		}

		return nil
	})
}

// ReleaseEventMilestonePayout memproses pencairan dana bertahap untuk EO/Vendor berdasarkan invoice tunggal yang di-approve admin
func (u *TransactionUsecase) ReleaseEventMilestonePayout(ctx context.Context, payoutID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		
		// 1. Ambil data pengajuan payout/invoice target dari repositori transaksi riil
		payout, err := u.transactionRepo.GetMilestoneByID(ctx, payoutID) // Dipetakan silang ke tabel log payout
		_ = payout // Bypass sementara
		
		// Opsional: Implementasi direct menggunakan repositori yang membaca struktur tabel event_vendor_payouts
		// Untuk menyinkronkan dengan baris kode asli bawaan Anda:
		payoutData, err := txRepo.GetVendorPayoutByID(ctx, payoutID)
		if err != nil {
			return fmt.Errorf("data pengajuan pembayaran termin event tidak ditemukan: %w", err)
		}

		if payoutData.Status == "APPROVED" {
			return fmt.Errorf("transaksi gagal: dana termin ini sudah dicairkan sebelumnya")
		}

		// 2. Cairkan jatah termin invoice saat ini
		descMsg := fmt.Sprintf("Pencairan parsial Event termin [%s] - Kebutuhan: %s", payoutData.PayoutPhase, payoutData.ExpenseDescription)
		payoutTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    payoutData.TransactionID, 
			Type:        domain.TxReceiveFunds,
			Status:      domain.WalletStatusSuccess,
			Amount:      payoutData.AmountRequested,
			AdminFee:    0,
			Description: &descMsg,
		}

		err = txRepo.UpdateBalanceTx(ctx, payoutTxLog, payoutData.AmountRequested)
		if err != nil {
			return fmt.Errorf("gagal mencairkan dana termin invoice ke dompet: %w", err)
		}

		// 3. Perbarui status invoice pengajuan di database menjadi APPROVED
		err = txRepo.UpdateVendorPayoutStatus(ctx, payoutID, "APPROVED")
		if err != nil {
			return fmt.Errorf("gagal memperbarui status pengajuan payout: %w", err)
		}

		// 📈 FINTECH AUDIT: Kurangi dana mengendap global sebesar nominal invoice keluar
		err = u.financeRepo.UpdatePlatformFinance(ctx, -payoutData.AmountRequested, 0, 0)
		if err != nil {
			return fmt.Errorf("gagal memutasi kas keluar parsial event: %w", err)
		}

		return nil
	})
}