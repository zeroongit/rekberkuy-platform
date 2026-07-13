package domain

import "time"

type Dispute struct {
	ID                string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID     string      `gorm:"type:uuid;not null;unique" json:"transaction_id"`
	Transaction       Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	RaisedBy          string      `gorm:"type:uuid;not null" json:"raised_by"`
	TargetPartyID     *string     `gorm:"type:uuid" json:"target_party_id,omitempty"`
	MediatorID        *string     `gorm:"type:uuid" json:"mediator_id,omitempty"`
	Raiser            UserProfile `gorm:"foreignKey:RaisedBy"`
	Target            UserProfile `gorm:"foreignKey:TargetPartyID"`
	Mediator          UserProfile `gorm:"foreignKey:MediatorID"`
	Reason            string      `gorm:"type:text;not null" json:"reason"`
	EvidenceURL       *string     `gorm:"type:text" json:"evidence_url,omitempty"`
	IsResolved        bool        `gorm:"type:boolean;not null;default:false" json:"is_resolved"`
	ResolutionSummary *string     `gorm:"type:text" json:"resolution_summary,omitempty"`
	CreatedAt         time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time   `gorm:"default:now()" json:"updated_at"`
}