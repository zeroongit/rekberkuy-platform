package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

type TransactionServicesUsecase struct {
	uow             domain.UnitOfWork
	transactionRepo domain.TransactionRepository
	financeCalc     *FinanceCalculator
	fraudClient     domain.FraudClient
	relayer         domain.Relayer
}

func NewTransactionServicesUsecase(uow domain.UnitOfWork, tr domain.TransactionRepository, fc *FinanceCalculator, fraud domain.FraudClient, relayer domain.Relayer) *TransactionServicesUsecase {
	return &TransactionServicesUsecase{
		uow:             uow,
		transactionRepo: tr,
		financeCalc:     fc,
		fraudClient:     fraud,
		relayer:         relayer,
	}
}

func (u *TransactionServicesUsecase) LockFundsServices(ctx context.Context, buyerID, sellerID string, amountBase int64, isRekberPay bool, sellerTier string, paymentMethod, idempotencyKey string, detail *domain.TransactionServices, milestones []domain.ServiceMilestone) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("service transaction amount must be greater than zero")
	}
	if detail == nil {
		return nil, errors.New("services transaction detail (category, deadline, brief) is required")
	}
	if detail.SubSubCategoryID == 0 || detail.BriefDescription == "" || detail.ProjectDeadline.IsZero() {
		return nil, errors.New("services detail requires sub_sub_category_id, project_deadline and brief_description")
	}
	if len(milestones) == 0 {
		return nil, errors.New("services transaction requires at least one milestone")
	}

	// Money-safety invariant: the milestone breakdown must exactly cover the
	// escrowed base amount, otherwise part of the escrow could become unreleasable.
	var milestoneTotal int64
	for i := range milestones {
		if milestones[i].Amount <= 0 {
			return nil, errors.New("each milestone amount must be greater than zero")
		}
		milestones[i].MilestoneIndex = i + 1
		milestones[i].Status = "PENDING"
		milestoneTotal += milestones[i].Amount
	}
	if milestoneTotal != amountBase {
		return nil, fmt.Errorf("milestone total %d must equal the transaction base amount %d", milestoneTotal, amountBase)
	}

	// Member cap: a regular USER may sell up to MaxMemberEventLimit per transaction.
	if err := enforceSellerLimit(ctx, u.uow, sellerID, amountBase); err != nil {
		return nil, err
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(domain.TypeServices, amountBase, isRekberPay, sellerTier)
	sellerFee := u.financeCalc.CalculateSellerServiceFee(domain.TypeServices, amountBase, sellerTier)

	amountGross := amountBase + buyerFee
	amountNet := amountBase - sellerFee
	midtransOrderID := fmt.Sprintf("%s%s", domain.OrderPrefixService, uuid.New().String()[:8])

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
	detail.TransactionID = txMaster.ID
	for i := range milestones {
		milestones[i].ID = uuid.New().String()
		milestones[i].TransactionID = txMaster.ID
	}

	// Master row + services detail + milestones commit atomically: releases
	// are keyed on milestone rows, so a master without its breakdown would
	// strand the escrow with nothing to release against.
	if err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		if err := stores.Transactions.CreateTransaction(ctx, txMaster); err != nil {
			return fmt.Errorf("failed to record service escrow transaction: %w", err)
		}
		return stores.Transactions.CreateServicesDetail(ctx, detail, milestones)
	}); err != nil {
		return nil, err
	}
	return txMaster, nil
}

// ConfirmPaymentServices settles a successful Midtrans payment for a services
// escrow transaction: it debits the buyer's wallet, moves the funds into escrow,
// and transitions the transaction WAITING_PAYMENT -> FUNDS_LOCKED. Idempotent —
// a replayed webhook for an already-locked transaction is a no-op success.
func (u *TransactionServicesUsecase) ConfirmPaymentServices(ctx context.Context, orderID string) error {
	return u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		tx, err := stores.Transactions.GetTransactionByMidtransOrderID(ctx, orderID)
		if err != nil {
			return err
		}
		if tx.Status != domain.StatusWaitingPayment {
			// Idempotent replay: Midtrans retries aggressively. If the payment was
			// already applied, acknowledge as a no-op success instead of erroring.
			return nil
		}

		descMsg := fmt.Sprintf("Payment success for services escrow transaction #%s", tx.ID)
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
			return fmt.Errorf("failed to change services transaction state to FUNDS_LOCKED: %w", err)
		}
		return stores.Finance.UpdatePlatformFinance(ctx, tx.AmountGross, 0, 0)
	})
}

func (u *TransactionServicesUsecase) ReleaseMilestoneFunds(ctx context.Context, milestoneID string) error {
	// 1. Pre-check (read outside tx) for fraud scoring.
	milestone, err := u.transactionRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to fetch milestone data: %w", err)
	}
	if milestone.Status == "RELEASED" {
		return fmt.Errorf("transaction rejected: this milestone fund has already been disbursed")
	}
	parentTx, err := u.transactionRepo.GetTransactionByID(ctx, milestone.TransactionID)
	if err != nil {
		return fmt.Errorf("parent transaction not found: %w", err)
	}

	// 2. Fraud scoring (fail-closed by default).
	if u.fraudClient != nil {
		_, isSafe, ferr := u.fraudClient.AnalyzeTransactionRisk(ctx, parentTx.BuyerID, milestone.Amount)
		if ferr != nil {
			return fmt.Errorf("milestone release refused: fraud screening unavailable: %w", ferr)
		}
		if !isSafe {
			return fmt.Errorf("milestone release refused: transaction flagged as high risk")
		}
	}

	// 3. ACID mutation.
	if err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		mLock, err := stores.Transactions.GetMilestoneByID(ctx, milestoneID)
		if err != nil {
			return fmt.Errorf("failed to fetch milestone data: %w", err)
		}
		if mLock.Status == "RELEASED" {
			return fmt.Errorf("transaction rejected: this milestone fund has already been disbursed")
		}
		txLock, err := stores.Transactions.GetTransactionByID(ctx, mLock.TransactionID)
		if err != nil {
			return fmt.Errorf("parent transaction not found: %w", err)
		}

		// Services commission is a flat 5% of the milestone amount (tier-independent);
		// it is retained by the platform as revenue, the seller receives the remainder.
		commission := u.financeCalc.CalculateSellerServiceFee(domain.TypeServices, mLock.Amount, "")
		netToSeller := mLock.Amount - commission

		descMsg := fmt.Sprintf("Service milestone fund disbursement [%s] for Transaction #%s", mLock.Title, txLock.ID)
		sellerTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    txLock.SellerID,
			Type:        domain.TxReceiveFunds,
			Status:      domain.WalletStatusSuccess,
			Amount:      netToSeller,
			AdminFee:    0,
			Description: &descMsg,
		}

		if err := stores.Wallets.UpdateBalanceTx(ctx, sellerTxLog, netToSeller); err != nil {
			return fmt.Errorf("failed to disburse milestone funds to freelancer: %w", err)
		}
		if err := stores.Transactions.UpdateMilestoneStatus(ctx, milestoneID, "RELEASED"); err != nil {
			return fmt.Errorf("failed to update milestone status in database: %w", err)
		}
		// The full milestone amount leaves escrow; the commission is recognised as revenue.
		return stores.Finance.UpdatePlatformFinance(ctx, -mLock.Amount, commission, 0)
	}); err != nil {
		return err
	}

	// 4. Audit log on-chain (best-effort, async).
	logAuditOnChain(u.relayer, u.transactionRepo, parentTx.ID, parentTx.BuyerID, parentTx.SellerID, milestone.Amount)
	return nil
}
