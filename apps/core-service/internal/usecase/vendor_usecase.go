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
	// 👈 KALIBRASI: Gunakan BusinessName dan Category sesuai file domain vendor.go asli Anda
	if vendor.BusinessName == "" || vendor.Category == "" {
		return fmt.Errorf("nama bisnis vendor dan kategori wajib diisi")
	}
	
	vendor.IsVerified = false // Tetap kunci false sebelum diverifikasi admin resmi
	
	return u.vendorRepo.CreateVendor(ctx, vendor)
}