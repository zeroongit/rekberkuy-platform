package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"rekberkuy/core-service/internal/domain"
)

func TestIdempotencyRepository_CheckOrLock_NewRequest(t *testing.T) {
	db, mock := newMockDB(t)
	r := &idempotencyRepository{db: db}
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO idempotency_records (id, request_path, response_body, response_status, created_at) VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT (id) DO NOTHING`)).
		WithArgs("key-1", "/path", []byte(nil), 0).WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected=1 -> new
	rec, isNew, err := r.CheckOrLock(context.Background(), minimalRecord("key-1", "/path"))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !isNew {
		t.Error("expected isNew=true for a new insert")
	}
	if rec.ID != "key-1" {
		t.Errorf("rec.ID = %s, want key-1", rec.ID)
	}
}

func TestIdempotencyRepository_CheckOrLock_Duplicate(t *testing.T) {
	db, mock := newMockDB(t)
	r := &idempotencyRepository{db: db}
	mock.ExpectExec(`INSERT INTO idempotency_records`).
		WillReturnResult(sqlmock.NewResult(0, 0)) // RowsAffected=0 -> conflict
	rows := sqlmock.NewRows([]string{"id", "request_path", "response_body", "response_status", "created_at"}).
		AddRow("key-1", "/path", []byte(`{"ok":true}`), 200, now)
	mock.ExpectQuery(`FROM idempotency_records WHERE id = \$1`).WithArgs("key-1").WillReturnRows(rows)

	rec, isNew, err := r.CheckOrLock(context.Background(), minimalRecord("key-1", "/path"))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if isNew {
		t.Error("expected isNew=false for a duplicate")
	}
	if rec.ResponseStatus != 200 {
		t.Errorf("status replay = %d, want 200", rec.ResponseStatus)
	}
}

func TestIdempotencyRepository_SaveResponse(t *testing.T) {
	db, mock := newMockDB(t)
	r := &idempotencyRepository{db: db}
	mock.ExpectExec(`UPDATE idempotency_records SET response_body = \$1, response_status = \$2 WHERE id = \$3`).
		WithArgs([]byte(`{"x":1}`), 201, "key-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.SaveResponse(context.Background(), "key-1", 201, []byte(`{"x":1}`)); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func minimalRecord(id, path string) *domain.IdempotencyRecord {
	return &domain.IdempotencyRecord{ID: id, RequestPath: path}
}
