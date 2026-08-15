package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"rekberkuy/core-service/internal/domain"
)

var (
	ErrWithdrawalAlreadyPaid  = errors.New("withdrawal request has already been marked paid")
	ErrInvalidWithdrawalReqt = errors.New("invalid withdrawal request")
)

// WithdrawalUsecase owns the wallet withdrawal flow:
//
//   - RequestWithdrawal debits the wallet (amount + WithdrawFeeToUser) and
//     records a PENDING withdrawal request in one UnitOfWork. The bank transfer
//     itself completes out-of-band (same stance as ADR-0002 for external
//     vendor disbursements).
//   - MarkWithdrawalDisbursed lets an admin confirm the transfer completed,
//     flipping the request to PAID.
type WithdrawalUsecase struct {
	uow                domain.UnitOfWork
	walletRepo         domain.WalletRepository
	disbursementClient domain.DisbursementClient
}

func NewWithdrawalUsecase(uow domain.UnitOfWork, wr domain.WalletRepository, dc domain.DisbursementClient) *WithdrawalUsecase {
	return &WithdrawalUsecase{uow: uow, walletRepo: wr, disbursementClient: dc}
}

// RequestWithdrawal creates a withdrawal request. Bank details are captured for
// the out-of-band transfer; the wallet is debited immediately (escrow-style:
// funds are reserved the moment the user asks for them).
//
// Fee model (true-up): the user pays a GROSS flat fee (domain.WithdrawFeeToUser,
// Rp7.500) that already covers the Midtrans disbursement cost. At request time
// the ESTIMATED cost (DisbursementClient.EstimateDisbursementFee — the
// configured MIDTRANS_DISBURSEMENT_FEE via the stub) is booked: revenue =
// gross − estimate, midtrans fees = estimate. When the disbursement is
// confirmed (MarkWithdrawalDisbursed) the REAL cost replaces the estimate and
// the ledger reconciles the difference automatically.
func (u *WithdrawalUsecase) RequestWithdrawal(ctx context.Context, userID string, amount int64, bankName, accountNumber, accountHolder string) (*domain.WithdrawalRequest, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be greater than zero", ErrInvalidWithdrawalReqt)
	}
	if bankName == "" || accountNumber == "" || accountHolder == "" {
		return nil, fmt.Errorf("%w: bank_name, account_number and account_holder are required", ErrInvalidWithdrawalReqt)
	}

	fee := int64(domain.WithdrawFeeToUser)

	// Estimated disbursement cost. When the real Midtrans Payout adapter lands,
	// this estimate comes live from the provider; the stub returns the
	// configured MIDTRANS_DISBURSEMENT_FEE. An estimation failure falls back
	// to the configured default — estimation must never block a withdrawal.
	estimate := int64(domain.WithdrawFeeMidtransCost)
	if u.disbursementClient != nil {
		req := domain.DisbursementRequest{
			BankName:      bankName,
			AccountNumber: accountNumber,
			AccountHolder: accountHolder,
			Amount:        amount,
		}
		if est, err := u.disbursementClient.EstimateDisbursementFee(ctx, req); err == nil && est >= 0 {
			estimate = est
		}
	}

	request := &domain.WithdrawalRequest{
		ID:            uuid.New().String(),
		UserID:        userID,
		Amount:        amount,
		Fee:           fee,
		MidtransCost:  estimate,
		BankName:      bankName,
		AccountNumber: accountNumber,
		AccountHolder: accountHolder,
		Status:        domain.WithdrawalPending,
	}

	if err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		descMsg := fmt.Sprintf("Withdrawal to %s account %s", bankName, accountNumber)
		walletTxLog := &domain.RekberPayTransaction{
			ID:          uuid.New().String(),
			WalletID:    userID,
			Type:        domain.TxWithdraw,
			Status:      domain.WalletStatusSuccess,
			Amount:      amount,
			AdminFee:    fee,
			Description: &descMsg,
		}
		// UpdateBalanceTx debits amount + AdminFee for TxWithdraw and refuses
		// when the balance is insufficient.
		if err := stores.Wallets.UpdateBalanceTx(ctx, walletTxLog, -amount); err != nil {
			return err
		}
		if err := stores.Wallets.CreateWithdrawalRequest(ctx, request); err != nil {
			return err
		}
		// Ledger split of the gross fee against the ESTIMATED cost; the true-up
		// at disbursement confirmation corrects any difference.
		return stores.Finance.UpdatePlatformFinance(ctx, 0, fee-estimate, estimate)
	}); err != nil {
		return nil, err
	}
	return request, nil
}

// ListMyWithdrawals returns the user's own withdrawal requests, newest first.
func (u *WithdrawalUsecase) ListMyWithdrawals(ctx context.Context, userID string, limit, offset int) ([]domain.WithdrawalRequest, error) {
	limit, offset = normalizePagination(limit, offset)
	return u.walletRepo.ListWithdrawalsByUser(ctx, userID, limit, offset)
}

// ListPendingWithdrawals returns the admin disbursement queue.
func (u *WithdrawalUsecase) ListPendingWithdrawals(ctx context.Context, limit, offset int) ([]domain.WithdrawalRequest, error) {
	limit, offset = normalizePagination(limit, offset)
	return u.walletRepo.ListPendingWithdrawals(ctx, limit, offset)
}

// MarkWithdrawalDisbursed lets an admin confirm the out-of-band bank transfer.
// actualCost is the REAL Midtrans disbursement fee for this transfer — when
// provided, the ledger true-up books the difference between the booked
// estimate and the real cost (midtrans fees += delta, revenue -= delta) in the
// same transaction as the status flip. A nil actualCost keeps the booked
// estimate (no reconciliation needed).
func (u *WithdrawalUsecase) MarkWithdrawalDisbursed(ctx context.Context, withdrawalID string, adminID string, actualCost *int64) error {
	if actualCost != nil && *actualCost < 0 {
		return fmt.Errorf("%w: actual midtrans cost cannot be negative", ErrInvalidWithdrawalReqt)
	}
	return u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		w, err := stores.Wallets.GetWithdrawalByID(ctx, withdrawalID)
		if err != nil {
			return err
		}
		if w.Status != domain.WithdrawalPending {
			return ErrWithdrawalAlreadyPaid
		}

		if actualCost != nil {
			if delta := *actualCost - w.MidtransCost; delta != 0 {
				// True-up: shift the difference between the booked estimate and
				// the real cost from revenue to the Midtrans fee ledger.
				if err := stores.Finance.UpdatePlatformFinance(ctx, 0, -delta, delta); err != nil {
					return fmt.Errorf("failed to reconcile withdrawal cost true-up: %w", err)
				}
			}
		}
		return stores.Wallets.MarkWithdrawalDisbursed(ctx, withdrawalID, adminID, actualCost)
	})
}
