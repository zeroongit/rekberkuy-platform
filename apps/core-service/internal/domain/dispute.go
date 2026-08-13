package domain

import (
	"context"
	"time"
)

// DisputeStatus is the lifecycle state of a dispute.
type DisputeStatus string

const (
	DisputeStatusOpen        DisputeStatus = "OPEN"         // Raised; awaiting an Admin to pick it up
	DisputeStatusUnderReview DisputeStatus = "UNDER_REVIEW" // An Admin has acknowledged and is mediating
	DisputeStatusResolved    DisputeStatus = "RESOLVED"     // Admin set a binding Outcome
)

// DisputeOutcome is the mediator's binding decision on a resolved dispute.
// It determines whether the Transaction ends REFUNDED or RELEASED.
type DisputeOutcome string

const (
	OutcomeRefundBuyer     DisputeOutcome = "REFUND_BUYER"      // Funds return to the buyer
	OutcomeReleaseToSeller DisputeOutcome = "RELEASE_TO_SELLER" // Funds pay out to the seller
)

// Dispute is a formal complaint raised by either party against a Transaction
// while its funds are locked, freezing the escrow pending Admin mediation.
// One Dispute per Transaction (unique TransactionID).
type Dispute struct {
	ID                string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID     string          `gorm:"type:uuid;not null;unique" json:"transaction_id"`
	Transaction       Transaction     `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	RaisedBy          string          `gorm:"type:uuid;not null" json:"raised_by"`
	TargetPartyID     *string         `gorm:"type:uuid" json:"target_party_id,omitempty"`
	MediatorID        *string         `gorm:"type:uuid" json:"mediator_id,omitempty"`
	Raiser            UserProfile     `gorm:"foreignKey:RaisedBy"`
	Target            UserProfile     `gorm:"foreignKey:TargetPartyID"`
	Mediator          UserProfile     `gorm:"foreignKey:MediatorID"`
	Reason            string          `gorm:"type:text;not null" json:"reason"`
	EvidenceURL       *string         `gorm:"type:text" json:"evidence_url,omitempty"`
	Status            DisputeStatus   `gorm:"type:varchar(50);not null;default:'OPEN'" json:"status"`
	Outcome           *DisputeOutcome `gorm:"type:varchar(50)" json:"outcome,omitempty"`
	IsResolved        bool            `gorm:"type:boolean;not null;default:false" json:"is_resolved"`
	ResolutionSummary *string         `gorm:"type:text" json:"resolution_summary,omitempty"`
	ResolvedAt        *time.Time      `json:"resolved_at,omitempty"`
	CreatedAt         time.Time       `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"default:now()" json:"updated_at"`
}

// DisputeRepository persists disputes. Methods that read during a resolution run
// inside the UnitOfWork (FOR UPDATE when transaction-bound) so concurrent
// mediation attempts cannot race on the same dispute.
type DisputeRepository interface {
	CreateDispute(ctx context.Context, d *Dispute) error
	GetDisputeByID(ctx context.Context, id string) (*Dispute, error)
	GetDisputeByTransactionID(ctx context.Context, transactionID string) (*Dispute, error)
	// AcknowledgeDispute moves a dispute OPEN -> UNDER_REVIEW.
	AcknowledgeDispute(ctx context.Context, id string) error
	// ResolveDispute finalizes a dispute: RESOLVED + outcome + mediator + summary.
	ResolveDispute(ctx context.Context, id string, adminID string, outcome DisputeOutcome, summary string) error
}
