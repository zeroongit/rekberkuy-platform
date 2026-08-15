package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

type TransactionEventsUsecase struct {
	uow             domain.UnitOfWork
	transactionRepo domain.TransactionRepository
	walletRepo      domain.WalletRepository // for pre-transaction reads (fraud check)
	financeCalc     *FinanceCalculator
	fraudClient     domain.FraudClient
	relayer         domain.Relayer
}

func NewTransactionEventsUsecase(uow domain.UnitOfWork, tr domain.TransactionRepository, wr domain.WalletRepository, fc *FinanceCalculator, fraud domain.FraudClient, relayer domain.Relayer) *TransactionEventsUsecase {
	return &TransactionEventsUsecase{
		uow:             uow,
		transactionRepo: tr,
		walletRepo:      wr,
		financeCalc:     fc,
		fraudClient:     fraud,
		relayer:         relayer,
	}
}

func (u *TransactionEventsUsecase) LockFundsEvents(ctx context.Context, buyerID, sellerID string, amountBase int64, isRekberPay bool, sellerTier string, paymentMethod, idempotencyKey string, detail *domain.TransactionEvents, allocations []domain.EventVendorAllocation) (*domain.Transaction, error) {
	if amountBase <= 0 {
		return nil, errors.New("event transaction amount must be greater than zero")
	}
	if detail == nil {
		return nil, errors.New("event transaction detail (category, event name, schedule) is required")
	}
	if detail.SubSubCategoryID == 0 || detail.EventName == "" || detail.EventStartTime.IsZero() || detail.EventEndTime.IsZero() {
		return nil, errors.New("event detail requires sub_sub_category_id, event_name, event_start_time and event_end_time")
	}
	if !detail.EventEndTime.After(detail.EventStartTime) {
		return nil, errors.New("event_end_time must be after event_start_time")
	}

	// Allocations are the EO's vendor pledges — they may be empty at lock time
	// (invoices drive the actual payouts), but when present they must be sane.
	var allocTotal int64
	for i := range allocations {
		if allocations[i].VendorID == "" {
			return nil, errors.New("each vendor allocation requires a vendor_id")
		}
		if allocations[i].AllocatedAmount <= 0 {
			return nil, errors.New("each vendor allocation amount must be greater than zero")
		}
		if allocations[i].Status == "" {
			allocations[i].Status = domain.VendorAllocationPledged
		}
		allocTotal += allocations[i].AllocatedAmount
	}
	if allocTotal > amountBase {
		return nil, fmt.Errorf("vendor allocation total %d exceeds the transaction base amount %d", allocTotal, amountBase)
	}

	// Member cap: a regular USER may sell up to MaxMemberEventLimit per transaction.
	if err := enforceSellerLimit(ctx, u.uow, sellerID, amountBase); err != nil {
		return nil, err
	}

	buyerFee := u.financeCalc.CalculateBuyerServiceFee(domain.TypeEvents, amountBase, isRekberPay, sellerTier)
	sellerFee := u.financeCalc.CalculateSellerServiceFee(domain.TypeEvents, amountBase, sellerTier)

	amountGross := amountBase + buyerFee
	amountNet := amountBase - sellerFee
	midtransOrderID := fmt.Sprintf("%s%s", domain.OrderPrefixEvent, uuid.New().String()[:8])

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
	detail.TransactionID = txMaster.ID
	for i := range allocations {
		allocations[i].ID = uuid.New().String()
		allocations[i].TransactionID = txMaster.ID
	}

	// Master row + event detail + vendor allocations commit atomically.
	if err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		if err := stores.Transactions.CreateTransaction(ctx, txMaster); err != nil {
			return fmt.Errorf("failed to record event escrow transaction: %w", err)
		}
		return stores.Transactions.CreateEventsDetail(ctx, detail, allocations)
	}); err != nil {
		return nil, err
	}
	return txMaster, nil
}

// SubmitEventVendorInvoice records a vendor's invoice (payment claim) against a
// locked event escrow. Only the event's organizer (the seller) may submit.
//
// Escrow cap: the sum of active vendor claims PLUS other vendors' unsettled
// pledges (PLEDGED allocations) may not exceed the escrow minus the 5% platform
// fee, so pledges + invoices together can never over-commit the escrow. The
// submitting vendor's own pledge is netted out — its invoice draws that pledge
// down rather than double-counting against the cap.
func (u *TransactionEventsUsecase) SubmitEventVendorInvoice(ctx context.Context, requesterID string, payout *domain.EventVendorPayout) error {
	if payout == nil {
		return errors.New("vendor invoice payload is required")
	}
	if payout.TransactionID == "" || payout.VendorName == "" || payout.VendorBankName == "" ||
		payout.VendorAccountNumber == "" || payout.ExpenseDescription == "" || payout.InvoiceFileURL == "" {
		return errors.New("vendor invoice requires transaction_id, vendor_name, vendor_bank_name, vendor_account_number, expense_description and invoice_file_url")
	}
	if payout.AmountRequested <= 0 {
		return errors.New("vendor invoice amount must be greater than zero")
	}

	return u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		tx, err := stores.Transactions.GetTransactionByID(ctx, payout.TransactionID)
		if err != nil {
			return err
		}
		if tx.Type != domain.TypeEvents {
			return fmt.Errorf("transaction %s is not an event escrow", tx.ID)
		}
		if requesterID != tx.SellerID {
			return errors.New("only the event organizer may submit vendor invoices for this event")
		}
		if tx.Status != domain.StatusFundsLocked {
			return fmt.Errorf("vendor invoices can only be submitted while the event escrow is FUNDS_LOCKED, current status %s", tx.Status)
		}

		activeTotal, err := stores.Transactions.GetActiveEventVendorPayoutsTotal(ctx, tx.ID)
		if err != nil {
			return err
		}

		// Pledge reservation: other vendors' unsettled pledges still hold a
		// promise on the escrow and must be honoured. The submitting vendor's
		// own pledge is excluded (netted) — this invoice claims against it
		// instead of double-counting.
		allocations, err := stores.Wallets.GetVendorAllocationsByTxID(ctx, tx.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch vendor allocations: %w", err)
		}
		var pledgeReservation int64
		for _, a := range allocations {
			if a.Status != domain.VendorAllocationPledged {
				continue
			}
			remaining := a.AllocatedAmount - a.ActualPaidAmount
			if remaining <= 0 {
				continue
			}
			if payout.VendorUserID != nil && *payout.VendorUserID == a.VendorID {
				continue
			}
			pledgeReservation += remaining
		}

		payableEscrow := tx.AmountGross - domain.EventPlatformFee(tx.AmountGross)
		totalClaims := activeTotal + pledgeReservation + payout.AmountRequested
		if totalClaims > payableEscrow {
			return fmt.Errorf("invoice rejected: total vendor claims %d (active invoices %d + reserved pledges %d + this invoice %d) would exceed the payable escrow %d (5%% platform fee reserved)",
				totalClaims, activeTotal, pledgeReservation, payout.AmountRequested, payableEscrow)
		}

		payout.ID = uuid.New().String()
		payout.Status = domain.VendorPayoutPending
		payout.IsDisbursedByMidtrans = false
		if payout.PayoutPhase == "" {
			payout.PayoutPhase = "FINAL_SETTLEMENT"
		}
		return stores.Wallets.CreateVendorPayoutRecord(ctx, payout)
	})
}

// ConfirmPaymentEvents settles a successful Midtrans payment for an event escrow
// transaction: it debits the buyer's (EO's) wallet, moves the funds into escrow,
// and transitions the transaction WAITING_PAYMENT -> FUNDS_LOCKED. Idempotent —
// a replayed webhook for an already-locked transaction is a no-op success.
func (u *TransactionEventsUsecase) ConfirmPaymentEvents(ctx context.Context, orderID string) error {
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

		descMsg := fmt.Sprintf("Payment success for event escrow transaction #%s", tx.ID)
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
			return fmt.Errorf("failed to change event transaction state to FUNDS_LOCKED: %w", err)
		}
		return stores.Finance.UpdatePlatformFinance(ctx, tx.AmountGross, 0, 0)
	})
}

// ProcessEventVendorPayouts disburses funds to all event vendors within a single
// transaction unit. Internal vendors (VendorUserID populated) are credited to their wallet;
// external vendors are marked PENDING_DISBURSEMENT (Midtrans disbursement in Phase 2).
// The parent transaction is released + audit logged on-chain only when all vendors are settled.
func (u *TransactionEventsUsecase) ProcessEventVendorPayouts(ctx context.Context, transactionID string) error {
	// 1. Pre-check + fraud scoring (outside the transactional boundary).
	tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.Status != domain.StatusFundsLocked {
		return fmt.Errorf("event funds failed to be split: transaction status must be FUNDS_LOCKED, current status %s", tx.Status)
	}
	if u.fraudClient != nil {
		_, isSafe, ferr := u.fraudClient.AnalyzeTransactionRisk(ctx, tx.BuyerID, tx.AmountGross)
		if ferr != nil {
			return fmt.Errorf("vendor payout refused: fraud screening unavailable: %w", ferr)
		}
		if !isSafe {
			return fmt.Errorf("vendor payout refused: transaction flagged as high risk")
		}
	}

	allSettled := false
	if err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		txLock, err := stores.Transactions.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return err
		}
		if txLock.Status != domain.StatusFundsLocked {
			return fmt.Errorf("event funds failed to be split: transaction status must be FUNDS_LOCKED, current status %s", txLock.Status)
		}

		payouts, err := stores.Transactions.GetEventVendorPayoutsByTxID(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("failed to fetch vendor invoice data: %w", err)
		}
		if len(payouts) == 0 {
			return fmt.Errorf("no vendor invoice data found for this event")
		}

		allSettled = true
		for _, payout := range payouts {
			if payout.Status == domain.VendorPayoutApproved {
				continue
			}

			if payout.VendorUserID != nil && *payout.VendorUserID != "" {
				descMsg := fmt.Sprintf("Full Event fund disbursement for Vendor: %s. Need: %s", payout.VendorName, payout.ExpenseDescription)
				vendorTxLog := &domain.RekberPayTransaction{
					ID:          uuid.New().String(),
					WalletID:    *payout.VendorUserID,
					Type:        domain.TxReceiveFunds,
					Status:      domain.WalletStatusSuccess,
					Amount:      payout.AmountRequested,
					AdminFee:    0,
					Description: &descMsg,
				}
				if err := stores.Wallets.UpdateBalanceTx(ctx, vendorTxLog, payout.AmountRequested); err != nil {
					return fmt.Errorf("failed to transfer funds to vendor %s: %w", payout.VendorName, err)
				}
			if err := stores.Transactions.UpdateEventVendorPayoutStatus(ctx, payout.ID, domain.VendorPayoutApproved); err != nil {
				return fmt.Errorf("failed to change vendor claim status %s: %w", payout.VendorName, err)
			}
			// The paid invoice settles (part of) the vendor's pledge — draw it
			// down so the reservation stops holding escrow it has already consumed.
			if err := stores.Wallets.MarkVendorAllocationClaimed(ctx, transactionID, *payout.VendorUserID, payout.AmountRequested); err != nil {
				return fmt.Errorf("failed to draw down vendor pledge %s: %w", payout.VendorName, err)
			}
			continue
			}

			// External vendor -> mark as waiting for Midtrans disbursement (Phase 2).
			if err := stores.Transactions.UpdateEventVendorPayoutStatus(ctx, payout.ID, domain.VendorPayoutPendingDisbursement); err != nil {
				return fmt.Errorf("failed to mark external vendor %s for disbursement: %w", payout.VendorName, err)
			}
			allSettled = false
		}

		if !allSettled {
			return nil
		}
		if err := stores.Transactions.UpdateTransactionStatus(ctx, transactionID, domain.StatusReleased); err != nil {
			return fmt.Errorf("failed to change event parent transaction state: %w", err)
		}
		// Settle the event ledger money-soundly:
		//   - vendors were just credited their AmountRequested (wallet mutations)
		//   - the platform takes a 5% fee (recognised as revenue)
		//   - the remaining surplus (AmountGross - vendor total - 5%) stays held in
		//     escrow, pending the EO-bonus / participant-refund distribution which
		//     needs a participant data model not yet present (see CalculateEventAudit).
		//     Previously the entire AmountGross left escrow with 0 revenue booked, so
		//     the surplus silently vanished — this keeps the ledger balanced.
		var vendorTotal int64
		for _, p := range payouts {
			vendorTotal += p.AmountRequested
		}
		platformFee := domain.EventPlatformFee(txLock.AmountGross)
		return stores.Finance.UpdatePlatformFinance(ctx, -(vendorTotal+platformFee), platformFee, txLock.MidtransFee)
	}); err != nil {
		return err
	}

	if allSettled {
		logAuditOnChain(u.relayer, u.transactionRepo, tx.ID, tx.BuyerID, tx.SellerID, tx.AmountGross)
	}
	return nil
}

// ReleaseEventMilestonePayout disburses a single event vendor invoice milestone.
func (u *TransactionEventsUsecase) ReleaseEventMilestonePayout(ctx context.Context, payoutID string) error {
	// 1. Pre-check + fraud scoring (outside the transactional boundary).
	payoutData, err := u.walletRepo.GetVendorPayoutByID(ctx, payoutID)
	if err != nil {
		return fmt.Errorf("event milestone payment request data not found: %w", err)
	}
	if payoutData.Status == domain.VendorPayoutApproved {
		return fmt.Errorf("transaction failed: this milestone fund has already been disbursed")
	}
	if u.fraudClient != nil && payoutData.VendorUserID != nil && *payoutData.VendorUserID != "" {
		_, isSafe, ferr := u.fraudClient.AnalyzeTransactionRisk(ctx, *payoutData.VendorUserID, payoutData.AmountRequested)
		if ferr != nil {
			return fmt.Errorf("vendor milestone release refused: fraud screening unavailable: %w", ferr)
		}
		if !isSafe {
			return fmt.Errorf("vendor milestone release refused: transaction flagged as high risk")
		}
	}

	credited := false
	if err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		pd, err := stores.Wallets.GetVendorPayoutByID(ctx, payoutID)
		if err != nil {
			return fmt.Errorf("event milestone payment request data not found: %w", err)
		}
		if pd.Status == domain.VendorPayoutApproved {
			return fmt.Errorf("transaction failed: this milestone fund has already been disbursed")
		}

		if pd.VendorUserID != nil && *pd.VendorUserID != "" {
			descMsg := fmt.Sprintf("Partial Event milestone disbursement [%s] - Need: %s", pd.PayoutPhase, pd.ExpenseDescription)
			payoutTxLog := &domain.RekberPayTransaction{
				ID:          uuid.New().String(),
				WalletID:    *pd.VendorUserID,
				Type:        domain.TxReceiveFunds,
				Status:      domain.WalletStatusSuccess,
				Amount:      pd.AmountRequested,
				AdminFee:    0,
				Description: &descMsg,
			}
			if err := stores.Wallets.UpdateBalanceTx(ctx, payoutTxLog, pd.AmountRequested); err != nil {
				return fmt.Errorf("failed to disburse invoice milestone funds to wallet: %w", err)
			}
		if err := stores.Wallets.UpdateVendorPayoutStatus(ctx, payoutID, domain.VendorPayoutApproved); err != nil {
			return fmt.Errorf("failed to update payout request status: %w", err)
		}
		// The paid invoice settles (part of) the vendor's pledge — draw it down.
		if err := stores.Wallets.MarkVendorAllocationClaimed(ctx, pd.TransactionID, *pd.VendorUserID, pd.AmountRequested); err != nil {
			return fmt.Errorf("failed to draw down vendor pledge: %w", err)
		}
		credited = true
		return stores.Finance.UpdatePlatformFinance(ctx, -pd.AmountRequested, 0, 0)
		}

		if err := stores.Wallets.UpdateVendorPayoutStatus(ctx, payoutID, domain.VendorPayoutPendingDisbursement); err != nil {
			return fmt.Errorf("failed to mark external vendor for disbursement: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if credited && payoutData.TransactionID != "" {
		if parent, err := u.transactionRepo.GetTransactionByID(ctx, payoutData.TransactionID); err == nil {
			logAuditOnChain(u.relayer, u.transactionRepo, parent.ID, parent.BuyerID, parent.SellerID, payoutData.AmountRequested)
		}
	}
	return nil
}
