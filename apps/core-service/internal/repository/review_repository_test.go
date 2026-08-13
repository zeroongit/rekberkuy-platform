package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"rekberkuy/core-service/internal/domain"
)

func TestReviewRepository_CreateReview(t *testing.T) {
	db, mock := newMockDB(t)
	r := &reviewRepository{db: db}
	comment := "great"
	mock.ExpectExec(`INSERT INTO reviews`).
		WithArgs("rev-1", "tx-1", "buyer-1", "seller-1", 5, &comment).
		WillReturnResult(sqlmock.NewResult(0, 1))
	review := &domain.Review{ID: "rev-1", TransactionID: "tx-1", ReviewerID: "buyer-1", RevieweeID: "seller-1", Rating: 5, Comment: &comment}
	if err := r.CreateReview(context.Background(), review); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestReviewRepository_GetReviewsByReviewee(t *testing.T) {
	db, mock := newMockDB(t)
	r := &reviewRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "transaction_id", "reviewer_id", "reviewee_id", "rating", "comment", "created_at"}).
		AddRow("rev-1", "tx-1", "buyer-1", "seller-1", 5, nil, now)
	mock.ExpectQuery(`FROM reviews WHERE reviewee_id = \$1`).WithArgs("seller-1").WillReturnRows(rows)
	out, err := r.GetReviewsByReviewee(context.Background(), "seller-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(out) != 1 || out[0].Rating != 5 {
		t.Errorf("unexpected reviews: %+v", out)
	}
}

func TestReviewRepository_GetAverageRatingForUser(t *testing.T) {
	db, mock := newMockDB(t)
	r := &reviewRepository{db: db}
	rows := sqlmock.NewRows([]string{"avg"}).AddRow(4.5)
	mock.ExpectQuery(`AVG\(rating\)`).WithArgs("seller-1").WillReturnRows(rows)
	avg, err := r.GetAverageRatingForUser(context.Background(), "seller-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if avg != 4.5 {
		t.Errorf("avg = %v, want 4.5", avg)
	}
}

func TestReviewRepository_GetAverageRatingForUser_None(t *testing.T) {
	db, mock := newMockDB(t)
	r := &reviewRepository{db: db}
	rows := sqlmock.NewRows([]string{"avg"}).AddRow(nil) // SQL NULL when a user has no reviews
	mock.ExpectQuery(`AVG\(rating\)`).WithArgs("nobody").WillReturnRows(rows)
	avg, err := r.GetAverageRatingForUser(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("expected success on no reviews, got: %v", err)
	}
	if avg != 0 {
		t.Errorf("avg = %v, want 0 when no reviews", avg)
	}
}
