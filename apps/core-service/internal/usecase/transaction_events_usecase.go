package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

type TransactionEventsUsecase struct {
	transactionRepo domain.TransactionRepository
	walletRepo      domain.WalletRepository
	financeRepo     domain.FinanceRepository
	financeCalc     *FinanceCalculator
}

func NewTransactionEventsUsecase(tr domain.TransactionRepository, wr domain.WalletRepository, fr domain.FinanceRepository, fc *FinanceCalculator) *TransactionEventsUsecase {
	return &TransactionEventsUsecase{
		transactionRepo: tr,
		walletRepo:      wr,
		financeRepo:     fr,
		financeCalc:     fc,
	}
}

func (u *TransactionEventsUsecase) LockFundsEvents(ctx context.Context, buyerID, sellerID string, amountBase int64, isRekberPay bool, sellerTier string, paymentMethod, idempotencyKey string) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("nominal transaksi event harus lebih besar dari nol")
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(domain.TypeEvents, amountBase, isRekberPay, sellerTier)
	sellerFee := u.financeCalc.CalculateSellerServiceFee(domain.TypeEvents, amountBase, sellerTier)

	amountGross := amountBase + buyerFee
	amountNet := amountBase - sellerFee
	midtransOrderID := fmt.Sprintf("REKBERKUY-EVENT-%s", uuid.New().String()[:8])

	txMaster := &domain.Transaction{
		ID:              uuid.New().String(),
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Type:            domain.TypeEvents,
		Status:          domain.StatusWaitingPayment,
		AmountBase:      amountBase,
		ShippingFee:     0,
		ServiceFee:      buyerFee,
		MidtransFee:     0,
		AmountGross:     amountGross,
		AmountNet:       amountNet,
		MidtransOrderID: midtransOrderID,
		IdempotencyKey:  idempotencyKey,
		PaymentMethod:   paymentMethod,
	}

	if err := u.transactionRepo.CreateTransaction(ctx, txMaster); err != nil {
		return nil, fmt.Errorf("gagal mencatat transaksi escrow event: %w", err)
	}
	return txMaster, nil
}

func (u *TransactionEventsUsecase) ProcessEventVendorPayouts(ctx context.Context, transactionID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}
		if tx.Status != domain.StatusFundsLocked {
			return fmt.Errorf("dana event gagal dipecah: status transaksi wajib FUNDS_LOCKED, status saat ini %s", tx.Status)
		}

		payouts, err := u.transactionRepo.GetEventVendorPayoutsByTxID(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("gagal mengambil data tagihan vendor: %w", err)
		}
		if len(payouts) == 0 {
			return fmt.Errorf("tidak ada data tagihan invoice vendor yang ditemukan untuk event ini")
		}

		for _, payout := range payouts {
			if payout.Status == "APPROVED" {
				continue
			}

			descMsg := fmt.Sprintf("Pencairan dana penuh Event untuk Vendor: %s. Kebutuhan: %s", payout.VendorName, payout.ExpenseDescription)
			vendorTxLog := &domain.RekberPayTransaction{
				ID:          uuid.New().String(),
				WalletID:    payout.TransactionID,
				Type:        domain.TxReceiveFunds,
				Status:      domain.WalletStatusSuccess,
				Amount:      payout.AmountRequested,
				AdminFee:    0,
				Description: &descMsg,
			}

			if err := txRepo.UpdateBalanceTx(ctx, vendorTxLog, payout.AmountRequested); err != nil {
				return fmt.Errorf("gagal mentransfer dana ke vendor %s: %w", payout.VendorName, err)
			}
			if err := u.transactionRepo.UpdateEventVendorPayoutStatus(ctx, payout.ID, "APPROVED"); err != nil {
				return fmt.Errorf("gagal merubah status klaim vendor %s: %w", payout.VendorName, err)
			}
		}

		if err := u.transactionRepo.UpdateTransactionStatus(ctx, transactionID, domain.StatusReleased); err != nil {
			return fmt.Errorf("gagal merubah state transaksi induk event: %w", err)
		}
		return u.financeRepo.UpdatePlatformFinance(ctx, -tx.AmountGross, tx.ServiceFee, tx.MidtransFee)
	})
}

func (u *TransactionEventsUsecase) ReleaseEventMilestonePayout(ctx context.Context, payoutID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		payoutData, err := txRepo.GetVendorPayoutByID(ctx, payoutID)
		if err != nil {
			return fmt.Errorf("data pengajuan pembayaran termin event tidak ditemukan: %w", err)
		}
		if payoutData.Status == "APPROVED" {
			return fmt.Errorf("transaksi gagal: dana termin ini sudah dicairkan sebelumnya")
		}

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

		if err := txRepo.UpdateBalanceTx(ctx, payoutTxLog, payoutData.AmountRequested); err != nil {
			return fmt.Errorf("gagal mencairkan dana termin invoice ke dompet: %w", err)
		}
		if err := txRepo.UpdateVendorPayoutStatus(ctx, payoutID, "APPROVED"); err != nil {
			return fmt.Errorf("gagal memperbarui status pengajuan payout: %w", err)
		}
		return u.financeRepo.UpdatePlatformFinance(ctx, -payoutData.AmountRequested, 0, 0)
	})
}