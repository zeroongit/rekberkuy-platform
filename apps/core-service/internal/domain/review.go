package domain

import (
	"context"
	"time"
)

// Review is a buyer's 1–5 rating (optional comment) left on the counterparty
// after a Transaction is RELEASED. One per (Transaction, Reviewer). Aggregated
// by the CRM worker into the merchant's average rating.
type Review struct {
	ID            string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID string      `gorm:"type:uuid;not null;uniqueIndex:idx_review_tx_reviewer,priority:1" json:"transaction_id"`
	Transaction   Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	ReviewerID    string      `gorm:"type:uuid;not null;uniqueIndex:idx_review_tx_reviewer,priority:2" json:"reviewer_id"`
	Reviewer      UserProfile `gorm:"foreignKey:ReviewerID"`
	RevieweeID    string      `gorm:"type:uuid;not null;index" json:"reviewee_id"`
	Reviewee      UserProfile `gorm:"foreignKey:RevieweeID"`
	Rating        int         `gorm:"type:integer;not null" json:"rating"`
	Comment       *string     `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt     time.Time   `gorm:"default:now()" json:"created_at"`
}

type ReviewRepository interface {
	CreateReview(ctx context.Context, review *Review) error
	GetReviewsByReviewee(ctx context.Context, revieweeID string) ([]Review, error)
	// GetAverageRatingForUser returns the mean rating a user has received across all
	// their reviews. The CRM worker uses this instead of a hardcoded value.
	GetAverageRatingForUser(ctx context.Context, userID string) (float64, error)
}
