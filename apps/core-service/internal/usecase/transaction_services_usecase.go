package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

type TransactionServicesUsecase struct {
	transactionRepo domain.TransactionRepository
	walletRepo      domain.WalletRepository
	financeRepo     domain.FinanceRepository
	financeCalc     *FinanceCalculator
}

func NewTransactionServicesUsecase(tr domain.TransactionRepository, wr domain.WalletRepository, fr domain.FinanceRepository, fc *FinanceCalculator) *TransactionServicesUsecase {
	return &TransactionServicesUsecase{
		transactionRepo: tr,
		walletRepo:      wr,
		financeRepo:     fr,
		financeCalc:     fc,
	}
}

func (u *TransactionServicesUsecase) LockFundsServices(ctx context.Context, buyerID, sellerID string, amountBase int64, isRekberPay bool, sellerTier string, paymentMethod, idempotencyKey string) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("nominal transaksi jasa harus lebih besar dari nol")
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(domain.TypeServices, amountBase, isRekberPay, sellerTier)
	sellerFee := u.financeCalc.CalculateSellerServiceFee(domain.TypeServices, amountBase, sellerTier)

	amountGross := amountBase + buyerFee
	amountNet := amountBase - sellerFee
	midtransOrderID := fmt.Sprintf("REKBERKUY-SERVICE-%s", uuid.New().String()[:8])

	txMaster := &domain.Transaction{
		ID:              uuid.New().String(),
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Type:            domain.TypeServices,
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
		return nil, fmt.Errorf("gagal mencatat transaksi escrow jasa: %w", err)
	}
	return txMaster, nil
}

func (u *TransactionServicesUsecase) ReleaseMilestoneFunds(ctx context.Context, milestoneID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		milestone, err := u.transactionRepo.GetMilestoneByID(ctx, milestoneID)
		if err != nil {
			return fmt.Errorf("gagal mengambil data milestone: %w", err)
		}
		if milestone.Status == "RELEASED" {
			return fmt.Errorf("transaksi ditolak: dana termin milestone ini sudah dicairkan sebelumnya")
		}

		tx, err := u.transactionRepo.GetTransactionByID(ctx, milestone.TransactionID)
		if err != nil {
			return fmt.Errorf("transaksi induk tidak ditemukan: %w", err)
		}

		descMsg := fmt.Sprintf("Pencairan dana milestone termin Jasa [%s] untuk Transaksi #%s", milestone.Title, tx.ID)
		sellerTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.SellerID,
			Type:        domain.TxReceiveFunds,
			Status:      domain.WalletStatusSuccess,
			Amount:      milestone.Amount,
			AdminFee:    0,
			Description: &descMsg,
		}

		if err := txRepo.UpdateBalanceTx(ctx, sellerTxLog, milestone.Amount); err != nil {
			return fmt.Errorf("gagal mencairkan dana termin milestone ke freelancer: %w", err)
		}
		if err := u.transactionRepo.UpdateMilestoneStatus(ctx, milestoneID, "RELEASED"); err != nil {
			return fmt.Errorf("gagal memperbarui status termin database: %w", err)
		}
		return u.financeRepo.UpdatePlatformFinance(ctx, -milestone.Amount, 0, 0)
	})
}