package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"rekberkuy/core-service/internal/domain"
)

func TestDisputeRepository_CreateDispute(t *testing.T) {
	db, mock := newMockDB(t)
	r := &disputeRepository{db: db}
	target := "seller-1"
	evidence := "https://evidence.example/img.png"
	mock.ExpectExec(`INSERT INTO disputes`).
		WithArgs("d-1", "tx-1", "buyer-1", &target, "item not as described", &evidence, domain.DisputeStatusOpen, false).
		WillReturnResult(sqlmock.NewResult(0, 1))
	d := &domain.Dispute{
		ID: "d-1", TransactionID: "tx-1", RaisedBy: "buyer-1", TargetPartyID: &target,
		Reason: "item not as described", EvidenceURL: &evidence,
		Status: domain.DisputeStatusOpen, IsResolved: false,
	}
	if err := r.CreateDispute(context.Background(), d); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestDisputeRepository_GetDisputeByID(t *testing.T) {
	db, mock := newMockDB(t)
	r := &disputeRepository{db: db}
	rows := sqlmock.NewRows([]string{"id", "transaction_id", "raised_by", "target_party_id", "mediator_id", "reason", "evidence_url", "status", "outcome", "is_resolved", "resolution_summary", "resolved_at", "created_at", "updated_at"}).
		AddRow("d-1", "tx-1", "buyer-1", nil, nil, "reason", nil, "OPEN", nil, false, nil, nil, now, now)
	mock.ExpectQuery(`FROM disputes WHERE id = \$1`).WithArgs("d-1").WillReturnRows(rows)

	d, err := r.GetDisputeByID(context.Background(), "d-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if d.Status != domain.DisputeStatusOpen {
		t.Errorf("status = %s, want OPEN", d.Status)
	}
}

func TestDisputeRepository_AcknowledgeDispute(t *testing.T) {
	db, mock := newMockDB(t)
	r := &disputeRepository{db: db}
	mock.ExpectExec(`UPDATE disputes`).
		WithArgs(domain.DisputeStatusUnderReview, "d-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.AcknowledgeDispute(context.Background(), "d-1"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestDisputeRepository_ResolveDispute(t *testing.T) {
	db, mock := newMockDB(t)
	r := &disputeRepository{db: db}
	mock.ExpectExec(`UPDATE disputes`).
		WithArgs(domain.DisputeStatusResolved, domain.OutcomeRefundBuyer, "admin-1", "buyer was right", "d-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := r.ResolveDispute(context.Background(), "d-1", "admin-1", domain.OutcomeRefundBuyer, "buyer was right"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}
