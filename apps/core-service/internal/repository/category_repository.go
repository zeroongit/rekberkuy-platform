package repository

import (
	"context"
	"database/sql"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

// categoryRepository is the read adapter for the seeded 3-tier taxonomy.
type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) list(ctx context.Context, table string, scan func(*sql.Rows) error) error {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, slug, created_at FROM `+table+` ORDER BY id ASC`)
	if err != nil {
		return fmt.Errorf("failed to list %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("failed to scan %s row: %w", table, err)
		}
	}
	return rows.Err()
}

func (r *categoryRepository) ListGoodsCategories(ctx context.Context) ([]domain.GoodsCategory, error) {
	out := []domain.GoodsCategory{}
	err := r.list(ctx, "goods_categories", func(rows *sql.Rows) error {
		var c domain.GoodsCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	return out, err
}

func (r *categoryRepository) ListServiceCategories(ctx context.Context) ([]domain.ServiceCategory, error) {
	out := []domain.ServiceCategory{}
	err := r.list(ctx, "service_categories", func(rows *sql.Rows) error {
		var c domain.ServiceCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	return out, err
}

func (r *categoryRepository) ListEventCategories(ctx context.Context) ([]domain.EventCategory, error) {
	out := []domain.EventCategory{}
	err := r.list(ctx, "event_categories", func(rows *sql.Rows) error {
		var c domain.EventCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	return out, err
}

// ListVendorSubCategories returns the vendor taxonomy's second tier
// (SOUND_SYSTEM, KATERING, ...) — the meaningful marketplace grouping.
func (r *categoryRepository) ListVendorSubCategories(ctx context.Context) ([]domain.VendorSubCategory, error) {
	query := `SELECT id, category_id, name, slug, created_at FROM vendor_sub_categories ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendor_sub_categories: %w", err)
	}
	defer rows.Close()

	out := []domain.VendorSubCategory{}
	for rows.Next() {
		var c domain.VendorSubCategory
		if err := rows.Scan(&c.ID, &c.CategoryID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan vendor_sub_categories row: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
