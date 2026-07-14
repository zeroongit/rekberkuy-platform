package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

type TransactionGoodsUsecase struct {
	transactionRepo domain.TransactionRepository
	walletRepo      domain.WalletRepository
	financeRepo     domain.FinanceRepository
	financeCalc     *FinanceCalculator
}

func NewTransactionGoodsUsecase(tr domain.TransactionRepository, wr domain.WalletRepository, fr domain.FinanceRepository, fc *FinanceCalculator) *TransactionGoodsUsecase {
	return &TransactionGoodsUsecase{
		transactionRepo: tr,
		walletRepo:      wr,
		financeRepo:     fr,
		financeCalc:     fc,
	}
}

func (u *TransactionGoodsUsecase) LockFundsGoods(ctx context.Context, buyerID, sellerID string, amountBase int64, isRekberPay bool, sellerTier string, shippingFee int64, paymentMethod, idempotencyKey string) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("nominal transaksi barang harus lebih besar dari nol")
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(domain.TypeGoods, amountBase, isRekberPay, sellerTier)
	sellerFee := u.financeCalc.CalculateSellerServiceFee(domain.TypeGoods, amountBase, sellerTier)

	amountGross := amountBase + buyerFee + shippingFee
	amountNet := amountBase - sellerFee
	midtransOrderID := fmt.Sprintf("REKBERKUY-GOODS-%s", uuid.New().String()[:8])

	txMaster := &domain.Transaction{
		ID:              uuid.New().String(),
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Type:            domain.TypeGoods,
		Status:          domain.StatusWaitingPayment,
		AmountBase:      amountBase,
		ShippingFee:     shippingFee,
		ServiceFee:      buyerFee,
		MidtransFee:     0,
		AmountGross:     amountGross,
		AmountNet:       amountNet,
		MidtransOrderID: midtransOrderID,
		IdempotencyKey:  idempotencyKey,
		PaymentMethod:   paymentMethod,
	}

	if err := u.transactionRepo.CreateTransaction(ctx, txMaster); err != nil {
		return nil, fmt.Errorf("gagal mencatat transaksi escrow barang: %w", err)
	}
	return txMaster, nil
}

func (u *TransactionGoodsUsecase) ConfirmPaymentGoods(ctx context.Context, transactionID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}
		if tx.Status != domain.StatusWaitingPayment {
			return fmt.Errorf("transaksi tidak dapat diproses: status saat ini adalah %s", tx.Status)
		}

		descMsg := fmt.Sprintf("Pembayaran sukses untuk transaksi escrow barang #%s", tx.ID)
		walletTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.BuyerID,
			Type:        domain.TxPayment,
			Status:      domain.WalletStatusSuccess,
			Amount:      tx.AmountGross,
			AdminFee:    tx.ServiceFee,
			Description: &descMsg,
		}

		if err := txRepo.UpdateBalanceTx(ctx, walletTxLog, -tx.AmountGross); err != nil {
			return fmt.Errorf("gagal mendebet saldo pembeli: %w", err)
		}
		if err := u.transactionRepo.UpdateTransactionStatus(ctx, tx.ID, domain.StatusFundsLocked); err != nil {
			return fmt.Errorf("gagal merubah state transaksi barang menjadi FUNDS_LOCKED: %w", err)
		}
		return u.financeRepo.UpdatePlatformFinance(ctx, tx.AmountGross, 0, 0)
	})
}

func (u *TransactionGoodsUsecase) ReleaseFundsGoods(ctx context.Context, transactionID string) error {
	return u.walletRepo.ExecuteInTransaction(ctx, func(txRepo domain.WalletRepository) error {
		tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}
		if tx.Status != domain.StatusFundsLocked {
			return fmt.Errorf("dana gagal dilepas: status transaksi wajib FUNDS_LOCKED, status saat ini %s", tx.Status)
		}

		descMsg := fmt.Sprintf("Penerimaan dana dari penyelesaian transaksi barang #%s", tx.ID)
		sellerTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.SellerID,
			Type:        domain.TxReceiveFunds,
			Status:      domain.WalletStatusSuccess,
			Amount:      tx.AmountNet,
			AdminFee:    0,
			Description: &descMsg,
		}

		if err := txRepo.UpdateBalanceTx(ctx, sellerTxLog, tx.AmountNet); err != nil {
			return fmt.Errorf("gagal mengredit saldo ke dompet penjual: %w", err)
		}
		if err := u.transactionRepo.UpdateTransactionStatus(ctx, tx.ID, domain.StatusReleased); err != nil {
			return fmt.Errorf("gagal merubah state transaksi menjadi RELEASED: %w", err)
		}
		return u.financeRepo.UpdatePlatformFinance(ctx, -tx.AmountGross, tx.ServiceFee, tx.MidtransFee)
	})
}