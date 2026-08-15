package usecase_test

import (
	"context"
	"errors"
	"testing"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// ============================================================================
// CATALOG USECASE — UNIT TESTS
// ============================================================================

func TestGetCategoryCatalog_ComposesAllDomains(t *testing.T) {
	ctx := context.Background()
	u := usecase.NewCatalogUsecase(&mockCategoryRepo{}, &mockVendorRepo{})

	catalog, err := u.GetCategoryCatalog(ctx)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(catalog.Goods) == 0 || len(catalog.Services) == 0 || len(catalog.Events) == 0 || len(catalog.Vendors) == 0 {
		t.Error("catalog must compose goods, services, events and vendor taxonomies")
	}
}

func TestGetCategoryCatalog_PropagatesError(t *testing.T) {
	ctx := context.Background()
	categoryRepo := &mockCategoryRepo{
		onListEvent: func(ctx context.Context) ([]domain.EventCategory, error) {
			return nil, errors.New("db down")
		},
	}
	u := usecase.NewCatalogUsecase(categoryRepo, &mockVendorRepo{})

	if _, err := u.GetCategoryCatalog(ctx); err == nil {
		t.Fatal("expected error when a taxonomy read fails")
	}
}

func TestListMarketplaceVendors_FiltersAndPaginates(t *testing.T) {
	ctx := context.Background()

	var gotCategory string
	var gotLimit, gotOffset int
	vendorRepo := &mockVendorRepo{
		onListVendors: func(ctx context.Context, category string, limit, offset int) ([]domain.VendorProfile, error) {
			gotCategory, gotLimit, gotOffset = category, limit, offset
			return []domain.VendorProfile{{VendorID: "v-1", BusinessName: "Katering Sedap"}}, nil
		},
	}
	u := usecase.NewCatalogUsecase(&mockCategoryRepo{}, vendorRepo)

	vendors, err := u.ListMarketplaceVendors(ctx, "CATERING", 500, -5)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(vendors) != 1 {
		t.Fatalf("vendors = %d, want 1", len(vendors))
	}
	if gotCategory != "CATERING" {
		t.Errorf("category = %s, want CATERING", gotCategory)
	}
	if gotLimit != 20 || gotOffset != 0 {
		t.Errorf("limit/offset = %d/%d, want clamped 20/0", gotLimit, gotOffset)
	}
}
