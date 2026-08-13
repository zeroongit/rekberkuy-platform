package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"rekberkuy/core-service/internal/domain"
)

// DisputeUsecase mediates contested escrow. Opening a dispute freezes a
// FUNDS_LOCKED transaction (FUNDS_LOCKED -> DISPUTED); resolving it moves money
// atomically inside the UnitOfWork — REFUND_BUYER credits the buyer (made whole,
// platform takes no fee) and ends the transaction REFUNDED; RELEASE_TO_SELLER
// credits the seller (AmountNet, normal commission) and ends it RELEASED.
// AI auto-resolution is deliberately deferred (ADR-0003); an Admin decides.
type DisputeUsecase struct {
	uow         domain.UnitOfWork
	disputeRepo domain.DisputeRepository
}

func NewDisputeUsecase(uow domain.UnitOfWork, disputeRepo domain.DisputeRepository) *DisputeUsecase {
	return &DisputeUsecase{uow: uow, disputeRepo: disputeRepo}
}

// OpenDispute raises a dispute on a FUNDS_LOCKED transaction, freezing it.
// Either the buyer or the seller may raise; the other party becomes the target.
func (u *DisputeUsecase) OpenDispute(ctx context.Context, transactionID, raisedBy, reason string, evidenceURL *string) (*domain.Dispute, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("a reason is required to open a dispute")
	}

	var created *domain.Dispute
	err := u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		tx, err := stores.Transactions.GetTransactionByID(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("failed to load transaction: %w", err)
		}
		if tx.Status != domain.StatusFundsLocked {
			return fmt.Errorf("a dispute may only be opened while funds are locked (current status: %s)", tx.Status)
		}

		var target string
		switch raisedBy {
		case tx.BuyerID:
			target = tx.SellerID
		case tx.SellerID:
			target = tx.BuyerID
		default:
			return errors.New("only the buyer or seller of a transaction may open a dispute")
		}

		d := &domain.Dispute{
			ID:            uuid.New().String(),
			TransactionID: transactionID,
			RaisedBy:      raisedBy,
			TargetPartyID: &target,
			Reason:        reason,
			EvidenceURL:   evidenceURL,
			Status:        domain.DisputeStatusOpen,
		}
		if err := stores.Disputes.CreateDispute(ctx, d); err != nil {
			return fmt.Errorf("failed to open dispute: %w", err)
		}
		if err := stores.Transactions.UpdateTransactionStatus(ctx, transactionID, domain.StatusDisputed); err != nil {
			return fmt.Errorf("failed to freeze transaction: %w", err)
		}
		created = d
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// AcknowledgeDispute moves a dispute OPEN -> UNDER_REVIEW (Admin picks it up).
func (u *DisputeUsecase) AcknowledgeDispute(ctx context.Context, disputeID string) error {
	return u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		d, err := stores.Disputes.GetDisputeByID(ctx, disputeID)
		if err != nil {
			return err
		}
		if d.Status != domain.DisputeStatusOpen {
			return fmt.Errorf("only an OPEN dispute can be acknowledged (current: %s)", d.Status)
		}
		return stores.Disputes.AcknowledgeDispute(ctx, disputeID)
	})
}

// ResolveDispute finalizes an UNDER_REVIEW dispute and moves escrow money
// atomically (dispute row + transaction status + wallet ledger in one UoW).
func (u *DisputeUsecase) ResolveDispute(ctx context.Context, disputeID, adminID string, outcome domain.DisputeOutcome, summary string) error {
	if outcome != domain.OutcomeRefundBuyer && outcome != domain.OutcomeReleaseToSeller {
		return errors.New("invalid dispute outcome")
	}

	return u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		d, err := stores.Disputes.GetDisputeByID(ctx, disputeID)
		if err != nil {
			return err
		}
		if d.Status != domain.DisputeStatusUnderReview {
			return fmt.Errorf("only an UNDER_REVIEW dispute can be resolved (current: %s)", d.Status)
		}

		tx, err := stores.Transactions.GetTransactionByID(ctx, d.TransactionID)
		if err != nil {
			return err
		}
		if tx.Status != domain.StatusDisputed {
			return fmt.Errorf("transaction is not in DISPUTED state (current: %s)", tx.Status)
		}

		switch outcome {
		case domain.OutcomeRefundBuyer:
			desc := fmt.Sprintf("Dispute refund to buyer for transaction #%s", tx.ID)
			refund := &domain.RekberPayTransaction{
				ID:          uuid.New().String(),
				WalletID:    tx.BuyerID,
				Type:        domain.TxRefund,
				Status:      domain.WalletStatusSuccess,
				Amount:      tx.AmountGross,
				Description: &desc,
			}
			if err := stores.Wallets.UpdateBalanceTx(ctx, refund, tx.AmountGross); err != nil {
				return fmt.Errorf("failed to refund buyer: %w", err)
			}
			if err := stores.Transactions.UpdateTransactionStatus(ctx, tx.ID, domain.StatusRefunded); err != nil {
				return fmt.Errorf("failed to mark transaction REFUNDED: %w", err)
			}
			// Escrow returns to buyer; the platform takes no fee on a disputed refund.
			if err := stores.Finance.UpdatePlatformFinance(ctx, -tx.AmountGross, 0, 0); err != nil {
				return fmt.Errorf("failed to reconcile finance: %w", err)
			}
		case domain.OutcomeReleaseToSeller:
			desc := fmt.Sprintf("Dispute release to seller for transaction #%s", tx.ID)
			payout := &domain.RekberPayTransaction{
				ID:          uuid.New().String(),
				WalletID:    tx.SellerID,
				Type:        domain.TxReceiveFunds,
				Status:      domain.WalletStatusSuccess,
				Amount:      tx.AmountNet,
				Description: &desc,
			}
			if err := stores.Wallets.UpdateBalanceTx(ctx, payout, tx.AmountNet); err != nil {
				return fmt.Errorf("failed to pay seller: %w", err)
			}
			if err := stores.Transactions.UpdateTransactionStatus(ctx, tx.ID, domain.StatusReleased); err != nil {
				return fmt.Errorf("failed to mark transaction RELEASED: %w", err)
			}
			// Same fee recognition as a normal release (commission kept by the platform).
			if err := stores.Finance.UpdatePlatformFinance(ctx, -tx.AmountGross, tx.ServiceFee, tx.MidtransFee); err != nil {
				return fmt.Errorf("failed to reconcile finance: %w", err)
			}
		}

		return stores.Disputes.ResolveDispute(ctx, disputeID, adminID, outcome, summary)
	})
}

// GetDispute fetches a dispute by id (read; no transactional lock needed).
func (u *DisputeUsecase) GetDispute(ctx context.Context, disputeID string) (*domain.Dispute, error) {
	return u.disputeRepo.GetDisputeByID(ctx, disputeID)
}
