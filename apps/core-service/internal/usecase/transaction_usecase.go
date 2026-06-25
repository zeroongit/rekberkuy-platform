package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid" 
	"rekberkuy/core-service/internal/domain"
)

type transactionUsecase struct {
	transactionRepo domain.TransactionRepository // SINKRONISASI: Injeksi Repo Transaksi Baru
	walletRepo      domain.WalletRepository
	financeCalc     *FinanceCalculator
}

// NewTransactionUsecase menginisialisasi manajer pengatur alur transaksi Rekberkuy
func NewTransactionUsecase(tr domain.TransactionRepository, wr domain.WalletRepository, fc *FinanceCalculator) *transactionUsecase {
	return &transactionUsecase{
		transactionRepo: tr,
		walletRepo:      wr,
		financeCalc:     fc,
	}
}

// LockFundsAwal menangani alur ketika pembeli mengunci dana mereka ke escrow Rekberkuy (Skenario Shopee Multi-Role)
func (u *transactionUsecase) LockFundsAwal(ctx context.Context, buyerID string, sellerID string, amountBase int64, rekberType domain.RekberType, isRekberPay bool, sellerTier string, shippingFee int64, paymentMethod string, idempotencyKey string) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("nominal transaksi harus lebih besar dari nol")
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(rekberType, amountBase, isRekberPay, sellerTier)
	
	// Hitung potongan komisi penjual saat pelepasan dana nanti
	sellerFee := u.financeCalc.CalculateSellerServiceFee(rekberType, amountBase, sellerTier)

	// Kalkulasi nominal total kotor (Gross) dan bersih (Net) sesuai kolom database Supabase
	amountGross := amountBase + buyerFee + shippingFee
	amountNet := amountBase - sellerFee

	midtransOrderID := fmt.Sprintf("REKBERKUY-ORDER-%s", uuid.New().String()[:8])


	txMaster := &domain.Transaction{
		ID:               uuid.New().String(),
		BuyerID:          buyerID,
		SellerID:         sellerID,
		Type:             rekberType,
		Status:           domain.StatusWaitingPayment, // State pertama: Menunggu Pembayaran Webhook Midtrans
		AmountBase:       amountBase,
		ShippingFee:      shippingFee,
		ServiceFee:       buyerFee,
		MidtransFee:      0,
		AmountGross:      amountGross,
		AmountNet:        amountNet,
		MidtransOrderID:  midtransOrderID,
		IdempotencyKey:   idempotencyKey,
		PaymentMethod:    paymentMethod,
		BlockchainTxHash: nil, // Masih kosong sebelum diproses secara gasless oleh Avalanche Relayer
	}
	
	err := u.transactionRepo.CreateTransaction(ctx, txMaster)
	if err != nil {
		return nil, fmt.Errorf("gagal mencatat transaksi master escrow: %w", err)
	}

	return txMaster, nil
}

// ConfirmPaymentWebhookMidtrans memproses perpindahan State Machine dari WAITING_PAYMENT ke FUNDS_LOCKED
func (u *transactionUsecase) ConfirmPaymentWebhookMidtrans(ctx context.Context, transactionID string) error {
	// Gunakan transaksi database ACID agar mutasi saldo pembeli dan update status transaksi terkunci rapat
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		
		// 1. Tarik data transaksi dan pasang kueri FOR UPDATE lock
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}

		if tx.Status != domain.StatusWaitingPayment {
			return fmt.Errorf("transaksi tidak dapat diproses: status saat ini adalah %s", tx.Status)
		}

		// 2. Potong saldo wallet RekberPay milik pembeli secara aman
		descMsg := fmt.Sprintf("Pembayaran sukses untuk transaksi escrow #%s", tx.ID)
		walletTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.BuyerID,
			Type:        domain.TxPayment, // SINKRONISASI: Menggunakan enum domain yang valid
			Status:      domain.WalletStatusSuccess,
			Amount:      tx.AmountGross,
			AdminFee:    tx.ServiceFee,
			Description: &descMsg,
		}

		err = txRepo.UpdateBalanceTx(ctx, walletTxLog, -tx.AmountGross)
		if err != nil {
			return fmt.Errorf("gagal mendebet saldo pembeli: %w", err)
		}

		// 3. Update status transaksi master menjadi FUNDS_LOCKED (Dana tersimpan di escrow platform)
		err = u.transactionRepo.UpdateTransactionStatus(ctx, tx.ID, domain.StatusFundsLocked)
		if err != nil {
			return fmt.Errorf("gagal merubah state transaksi menjadi FUNDS_LOCKED: %w", err)
		}

		return nil
	})
}

// ReleaseFundsSelesai memproses pelepasan dana escrow dari platform ke dompet milik penjual (Barang / Jasa)
func (u *transactionUsecase) ReleaseFundsSelesai(ctx context.Context, transactionID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}

		if tx.Status != domain.StatusFundsLocked {
			return fmt.Errorf("dana gagal dilepas: status transaksi wajib FUNDS_LOCKED, status saat ini %s", tx.Status)
		}

		// 1. Tambahkan saldo bersih (AmountNet) ke dompet wallet milik penjual
		descMsg := fmt.Sprintf("Penerimaan dana dari penyelesaian transaksi escrow #%s", tx.ID)
		sellerTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.SellerID,
			Type:        domain.TxReceiveFunds, // SINKRONISASI: Menggunakan enum domain yang valid
			Status:      domain.WalletStatusSuccess,
			Amount:      tx.AmountNet,
			AdminFee:    0,
			Description: &descMsg,
		}

		err = txRepo.UpdateBalanceTx(ctx, sellerTxLog, tx.AmountNet)
		if err != nil {
			return fmt.Errorf("gagal mengredit saldo ke dompet penjual: %w", err)
		}

		// 2. Ubah status transaksi menjadi RELEASED
		err = u.transactionRepo.UpdateTransactionStatus(ctx, tx.ID, domain.StatusReleased)
		if err != nil {
			return fmt.Errorf("gagal merubah state transaksi menjadi RELEASED: %w", err)
		}

		return nil
	})
}