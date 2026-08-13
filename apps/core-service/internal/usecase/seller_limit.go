package usecase

import (
	"context"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

// enforceSellerLimit enforces the member (USER-role) selling cap.
//
// A plain USER may sell goods, services, or events up to
// domain.MaxMemberEventLimit (Rp 10.000.000) per transaction. Verified
// merchants, vendors, and event organisers are unrestricted.
//
// The seller's profile is only loaded when the amount exceeds the cap (cheap
// fast path), so the common small transaction never touches the database here.
func enforceSellerLimit(ctx context.Context, uow domain.UnitOfWork, sellerID string, amount int64) error {
	if amount <= domain.MaxMemberEventLimit {
		return nil
	}

	var seller *domain.UserProfile
	if err := uow.Do(ctx, func(ctx context.Context, stores domain.TxStores) error {
		u, err := stores.Users.GetProfileByID(ctx, sellerID)
		if err != nil {
			return fmt.Errorf("failed to verify seller eligibility: %w", err)
		}
		seller = u
		return nil
	}); err != nil {
		return err
	}

	if seller.Role == domain.RoleUser {
		return fmt.Errorf(
			"transaction rejected: regular users may sell up to Rp %d per transaction (requested Rp %d); complete seller verification to raise the limit",
			domain.MaxMemberEventLimit, amount,
		)
	}
	return nil
}
