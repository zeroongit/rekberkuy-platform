package usecase

import (
	"context"

	"rekberkuy/core-service/internal/domain"
)

// CatalogUsecase serves the public catalog reads: the 3-tier category taxonomy
// and the vendor marketplace browse.
type CatalogUsecase struct {
	categoryRepo domain.CategoryRepository
	vendorRepo   domain.VendorRepository
}

func NewCatalogUsecase(cr domain.CategoryRepository, vr domain.VendorRepository) *CatalogUsecase {
	return &CatalogUsecase{categoryRepo: cr, vendorRepo: vr}
}

type CategoryCatalog struct {
	Goods    []domain.GoodsCategory       `json:"goods"`
	Services []domain.ServiceCategory     `json:"services"`
	Events   []domain.EventCategory       `json:"events"`
	Vendors  []domain.VendorSubCategory   `json:"vendors"`
}

// GetCategoryCatalog returns the top-level taxonomy across all four domains.
func (u *CatalogUsecase) GetCategoryCatalog(ctx context.Context) (*CategoryCatalog, error) {
	goods, err := u.categoryRepo.ListGoodsCategories(ctx)
	if err != nil {
		return nil, err
	}
	services, err := u.categoryRepo.ListServiceCategories(ctx)
	if err != nil {
		return nil, err
	}
	events, err := u.categoryRepo.ListEventCategories(ctx)
	if err != nil {
		return nil, err
	}
	vendors, err := u.categoryRepo.ListVendorSubCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &CategoryCatalog{Goods: goods, Services: services, Events: events, Vendors: vendors}, nil
}

// ListMarketplaceVendors returns vendors, optionally filtered by category.
func (u *CatalogUsecase) ListMarketplaceVendors(ctx context.Context, category string, limit, offset int) ([]domain.VendorProfile, error) {
	limit, offset = normalizePagination(limit, offset)
	return u.vendorRepo.ListVendors(ctx, category, limit, offset)
}
