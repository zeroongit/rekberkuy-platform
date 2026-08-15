package usecase

import (
	"context"
	"errors"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

var ErrNotTransactionParty = errors.New("you are not a party to this transaction")

// TransactionQueryUsecase serves the read APIs: a user's transaction list and
// the composed detail view (master + type-specific container + milestones /
// vendor payouts). Read-only — never moves money.
type TransactionQueryUsecase struct {
	transactionRepo domain.TransactionRepository
}

func NewTransactionQueryUsecase(tr domain.TransactionRepository) *TransactionQueryUsecase {
	return &TransactionQueryUsecase{transactionRepo: tr}
}

// TransactionDetailView is the composed read model for the detail endpoint.
type TransactionDetailView struct {
	domain.Transaction
	Goods        *domain.TransactionGoods     `json:"goods,omitempty"`
	Services     *domain.TransactionServices  `json:"services,omitempty"`
	Milestones   []domain.ServiceMilestone    `json:"milestones,omitempty"`
	Events       *domain.TransactionEvents    `json:"events,omitempty"`
	VendorPayout []domain.EventVendorPayout   `json:"vendor_payouts,omitempty"`
}

// ListUserTransactions returns transactions where the user is buyer or seller.
func (u *TransactionQueryUsecase) ListUserTransactions(ctx context.Context, userID string, limit, offset int) ([]domain.Transaction, error) {
	limit, offset = normalizePagination(limit, offset)
	return u.transactionRepo.ListTransactionsByUser(ctx, userID, limit, offset)
}

// GetTransactionDetail composes the master row with its type-specific detail.
// requesterID must be the buyer or the seller of the transaction.
func (u *TransactionQueryUsecase) GetTransactionDetail(ctx context.Context, requesterID string, transactionID string) (*TransactionDetailView, error) {
	tx, err := u.transactionRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if requesterID != tx.BuyerID && requesterID != tx.SellerID {
		return nil, ErrNotTransactionParty
	}

	view := &TransactionDetailView{Transaction: *tx}
	switch tx.Type {
	case domain.TypeGoods:
		if view.Goods, err = u.transactionRepo.GetGoodsDetailByTxID(ctx, tx.ID); err != nil {
			return nil, fmt.Errorf("failed to load goods detail: %w", err)
		}
	case domain.TypeServices:
		if view.Services, err = u.transactionRepo.GetServicesDetailByTxID(ctx, tx.ID); err != nil {
			return nil, fmt.Errorf("failed to load services detail: %w", err)
		}
		if view.Milestones, err = u.transactionRepo.GetMilestonesByTxID(ctx, tx.ID); err != nil {
			return nil, fmt.Errorf("failed to load milestones: %w", err)
		}
	case domain.TypeEvents:
		if view.Events, err = u.transactionRepo.GetEventsDetailByTxID(ctx, tx.ID); err != nil {
			return nil, fmt.Errorf("failed to load event detail: %w", err)
		}
		if view.VendorPayout, err = u.transactionRepo.GetEventVendorPayoutsByTxID(ctx, tx.ID); err != nil {
			return nil, fmt.Errorf("failed to load vendor payouts: %w", err)
		}
	}
	return view, nil
}
