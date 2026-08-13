package usecase

import (
	"context"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type VendorUsecase struct {
	vendorRepo domain.VendorRepository
}

func NewVendorUsecase(vr domain.VendorRepository) *VendorUsecase {
	return &VendorUsecase{vendorRepo: vr}
}

func (u *VendorUsecase) RegisterVendorProfile(ctx context.Context, vendor *domain.VendorProfile) error {
	// 👈 CALIBRATION: Use BusinessName and Category per your actual domain vendor.go file
	if vendor.BusinessName == "" || vendor.Category == "" {
		return fmt.Errorf("vendor business name and category are required")
	}

	vendor.IsVerified = false // Keep locked false until officially verified by admin

	return u.vendorRepo.CreateVendor(ctx, vendor)
}
