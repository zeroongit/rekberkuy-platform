package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"rekberkuy/core-service/internal/domain"
)

// TopUpWallet validates the amount, records a PENDING transaction draft, then creates
// a Snap Transaction in Midtrans. Returns the draft + snap token/redirect URL.
// The balance increases when the Midtrans webhook confirms settlement (ConfirmTopUp).
func (u *UserUsecase) TopUpWallet(ctx context.Context, userID string, amount int64) (*domain.RekberPayTransaction, *domain.SnapResult, error) {
	if amount <= 0 {
		return nil, nil, fmt.Errorf("top-up amount must be greater than Rp 0")
	}
	if u.midtrans == nil {
		return nil, nil, fmt.Errorf("payment gateway not configured")
	}

	orderID := fmt.Sprintf("%s%s", domain.OrderPrefixTopUp, uuid.New().String()[:8])
	descMsg := fmt.Sprintf("Top-up balance request via Midtrans of Rp %d (order %s)", amount, orderID)

	// 1. PENDING transaction draft (delta 0). midtrans_topup_id = orderID for webhook lookup.
	draft := &domain.RekberPayTransaction{
		ID:              uuid.New().String(),
		WalletID:        userID,
		Type:            domain.TxTopUp,
		Status:          domain.WalletStatusPending,
		Amount:          amount,
		MidtransTopUpID: &orderID,
		Description:     &descMsg,
	}
	if err := u.walletRepo.UpdateBalanceTx(ctx, draft, 0); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize top-up transaction: %w", err)
	}

	// 2. Create Snap transaction in Midtrans (server-to-server).
	snap, err := u.midtrans.CreateSnapTransaction(ctx, domain.SnapRequest{
		OrderID:     orderID,
		GrossAmount: amount,
	})
	if err != nil {
		// Snap failed -> mark draft FAILED (balance unchanged because delta is 0).
		_ = u.walletRepo.MarkWalletTxStatusByOrderID(ctx, orderID, domain.WalletStatusFailed)
		return nil, nil, fmt.Errorf("failed to create snap transaction: %w", err)
	}

	return draft, &snap, nil
}

// MarkTopUpFailed marks the top-up draft as FAILED (payment deny/expire/cancel).
func (u *UserUsecase) MarkTopUpFailed(ctx context.Context, orderID string) error {
	return u.walletRepo.MarkWalletTxStatusByOrderID(ctx, orderID, domain.WalletStatusFailed)
}

// ConfirmTopUp is called by the Midtrans webhook when a top-up payment settles/captures.
// It credits the user's balance by the amount and marks the draft SUCCESS. Idempotent:
// if the draft is already SUCCESS, it does not credit twice.
func (u *UserUsecase) ConfirmTopUp(ctx context.Context, orderID string, midtransTxID string) error {
	return u.uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		draft, err := stores.Wallets.GetWalletTxByMidtransOrderID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("top-up draft not found: %w", err)
		}
		if draft.Status == domain.WalletStatusSuccess {
			return nil // idempotent: already credited previously
		}

		creditTxID := uuid.New().String()
		topUpRef := midtransTxID
		creditDesc := fmt.Sprintf("Top-up confirmation succeeded for order %s", orderID)
		credit := &domain.RekberPayTransaction{
			ID:              creditTxID,
			WalletID:        draft.WalletID,
			Type:            domain.TxTopUp,
			Status:          domain.WalletStatusSuccess,
			Amount:          draft.Amount,
			MidtransTopUpID: &topUpRef,
			Description:     &creditDesc,
		}
		// Credit balance (+amount).
		if err := stores.Wallets.UpdateBalanceTx(ctx, credit, draft.Amount); err != nil {
			return fmt.Errorf("failed to credit top-up balance: %w", err)
		}
		// Mark the initial draft as SUCCESS.
		if err := stores.Wallets.MarkWalletTxStatusByOrderID(ctx, orderID, domain.WalletStatusSuccess); err != nil {
			return fmt.Errorf("failed to mark top-up draft SUCCESS: %w", err)
		}
		return nil
	})
}
