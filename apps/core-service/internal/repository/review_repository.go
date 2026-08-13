package repository

import (
	"context"
	"database/sql"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

type reviewRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewReviewRepository returns the database adapter for merchant reviews.
func NewReviewRepository(db *sql.DB) domain.ReviewRepository {
	return &reviewRepository{db: db}
}

// CreateReview stores a buyer's review. The (transaction_id, reviewer_id)
// composite unique index enforces "one review per transaction per reviewer".
func (r *reviewRepository) CreateReview(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, transaction_id, reviewer_id, reviewee_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	args := []any{review.ID, review.TransactionID, review.ReviewerID, review.RevieweeID, review.Rating, review.Comment}
	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, args...)
	} else {
		_, err = r.db.ExecContext(ctx, query, args...)
	}
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}
	return nil
}

func (r *reviewRepository) GetReviewsByReviewee(ctx context.Context, revieweeID string) ([]domain.Review, error) {
	query := `SELECT id, transaction_id, reviewer_id, reviewee_id, rating, comment, created_at FROM reviews WHERE reviewee_id = $1 ORDER BY created_at DESC`
	var rows *sql.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query, revieweeID)
	} else {
		rows, err = r.db.QueryContext(ctx, query, revieweeID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews: %w", err)
	}
	defer rows.Close()

	var out []domain.Review
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.TransactionID, &rv.ReviewerID, &rv.RevieweeID, &rv.Rating, &rv.Comment, &rv.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		out = append(out, rv)
	}
	return out, nil
}

// GetAverageRatingForUser returns the mean rating a user has received. With no
// reviews, AVG returns SQL NULL, which scans to 0 — the caller treats 0 as
// "use the neutral default".
func (r *reviewRepository) GetAverageRatingForUser(ctx context.Context, userID string) (float64, error) {
	query := `SELECT AVG(rating) FROM reviews WHERE reviewee_id = $1`
	var avg sql.NullFloat64
	var err error
	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, userID).Scan(&avg)
	} else {
		err = r.db.QueryRowContext(ctx, query, userID).Scan(&avg)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to compute average rating: %w", err)
	}
	return avg.Float64, nil
}
