package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

type TransactionGoodsUsecase struct {
	uow             domain.UnitOfWork
	transactionRepo domain.TransactionRepository
	financeCalc     *FinanceCalculator
	fraudClient     domain.FraudClient
	relayer         domain.Relayer
}

func NewTransactionGoodsUsecase(uow domain.UnitOfWork, tr domain.TransactionRepository, fc *FinanceCalculator, fraud domain.FraudClient, relayer domain.Relayer) *TransactionGoodsUsecase {
	return &TransactionGoodsUsecase{
		uow:             uow,
		transactionRepo: tr,
		financeCalc:     fc,
		fraudClient:     fraud,
		relayer:         relayer,
	}
}

func (u *TransactionGoodsUsecase) LockFundsGoods(ctx context.Context, buyerID, sellerID string, amountBase int64, isRekberPay bool, sellerTier string, shippingFee int64, paymentMethod, idempotencyKey string) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("goods transaction amount must be greater than zero")
	}

	// Member cap: a regular USER may sell up to MaxMemberEventLimit per transaction.
	if err := enforceSellerLimit(ctx, u.uow, sellerID, amountBase); err != nil {
		return nil, err
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(domain.TypeGoods, amountBase, isRekberPay, sellerTier)
	sellerFee := u.financeCalc.CalculateSellerServiceFee(domain.TypeGoods, amountBase, sellerTier)

	amountGross := amountBase + buyerFee + shippingFee
	amountNet := amountBase - sellerFee
	midtransOrderID := fmt.Sprintf("%s%s", domain.OrderPrefixGoods, uuid.New().String()[:8])

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
		return nil, fmt.Errorf("failed to record goods escrow transaction: %w", err)
	}
	return txMaster, nil
}

func (u *TransactionGoodsUsecase) ConfirmPaymentGoods(ctx context.Context, orderID string) error {
	return u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		tx, err := stores.Transactions.GetTransactionByMidtransOrderID(ctx, orderID)
		if err != nil {
			return err
		}
		if tx.Status != domain.StatusWaitingPayment {
			// Idempotent replay: Midtrans retries webhooks aggressively. If the payment
			// was already applied (FUNDS_LOCKED or any later state), acknowledge as a
			// no-op success instead of erroring (which would trigger more retries).
			return nil
		}

		descMsg := fmt.Sprintf("Payment success for goods escrow transaction #%s", tx.ID)
		walletTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    tx.BuyerID,
			Type:        domain.TxPayment,
			Status:      domain.WalletStatusSuccess,
			Amount:      tx.AmountGross,
			AdminFee:    tx.ServiceFee,
			Description: &descMsg,
		}

		if err := stores.Wallets.UpdateBalanceTx(ctx, walletTxLog, -tx.AmountGross); err != nil {
			return fmt.Errorf("failed to debit buyer balance: %w", err)
		}
		if err := stores.Transactions.UpdateTransactionStatus(ctx, tx.ID, domain.StatusFundsLocked); err != nil {
			return fmt.Errorf("failed to change goods transaction state to FUNDS_LOCKED: %w", err)
		}
		return stores.Finance.UpdatePlatformFinance(ctx, tx.AmountGross, 0, 0)
	})
}

func (u *TransactionGoodsUsecase) ReleaseFundsGoods(ctx context.Context, transactionID string) error {
	// 1. Pre-check (read outside the transactional boundary) for fraud evaluation.
	tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.Status != domain.StatusFundsLocked {
		return fmt.Errorf("funds failed to be released: transaction status must be FUNDS_LOCKED, current status %s", tx.Status)
	}

	// 2. Fraud scoring (fail-closed by default: if screening is unavailable or
	// flags risk, refuse the release — money safety over availability).
	if u.fraudClient != nil {
		_, isSafe, ferr := u.fraudClient.AnalyzeTransactionRisk(ctx, tx.BuyerID, tx.AmountGross)
		if ferr != nil {
			return fmt.Errorf("release of funds %s refused: fraud screening unavailable: %w", tx.ID, ferr)
		}
		if !isSafe {
			return fmt.Errorf("release of funds %s refused: transaction flagged as high risk", tx.ID)
		}
	}

	// 3. ACID mutation: credit seller + RELEASED state + cash reconciliation.
	if err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		txLock, err := stores.Transactions.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}
		if txLock.Status != domain.StatusFundsLocked {
			return fmt.Errorf("funds failed to be released: transaction status must be FUNDS_LOCKED, current status %s", txLock.Status)
		}

		descMsg := fmt.Sprintf("Receipt of funds from goods transaction settlement #%s", txLock.ID)
		sellerTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    txLock.SellerID,
			Type:        domain.TxReceiveFunds,
			Status:      domain.WalletStatusSuccess,
			Amount:      txLock.AmountNet,
			AdminFee:    0,
			Description: &descMsg,
		}

		if err := stores.Wallets.UpdateBalanceTx(ctx, sellerTxLog, txLock.AmountNet); err != nil {
			return fmt.Errorf("failed to credit funds to seller wallet: %w", err)
		}
		if err := stores.Transactions.UpdateTransactionStatus(ctx, txLock.ID, domain.StatusReleased); err != nil {
			return fmt.Errorf("failed to change transaction state to RELEASED: %w", err)
		}
		// Recognise the full platform retention (buyer-protection fee + seller
		// commission) as revenue so the ledger balances: AmountGross leaves escrow,
		// AmountNet goes to the seller, the remainder is platform revenue.
		return stores.Finance.UpdatePlatformFinance(ctx, -txLock.AmountGross, txLock.AmountGross-txLock.AmountNet, txLock.MidtransFee)
	}); err != nil {
		return err
	}

	// 4. Audit log on-chain (best-effort, async) — only hash + amount, no PII.
	logAuditOnChain(u.relayer, u.transactionRepo, tx.ID, tx.BuyerID, tx.SellerID, tx.AmountGross)
	return nil
}
