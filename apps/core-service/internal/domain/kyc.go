package domain

import "time"

type KYCStatus string

const (
	KYCPending  KYCStatus = "PENDING"
	KYCApproved KYCStatus = "APPROVED"
	KYCRejected KYCStatus = "REJECTED"
)

type KYCSubmission struct {
	ID           string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string       `gorm:"type:uuid;not null;unique;index" json:"user_id"`
	UserProfile  UserProfile  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	TargetRole   UserRole     `gorm:"type:varchar(50);not null" json:"target_role"` // VERIFIED_MERCHANT, VERIFIED_VENDOR, EVENT_ORGANIZER
	IDCardNumber string       `gorm:"type:varchar(50);not null;unique" json:"id_card_number"`
	IDCardURL    string       `gorm:"type:text;not null" json:"id_card_url"`
	SelfieURL    string       `gorm:"type:text;not null" json:"selfie_url"`
	Status       KYCStatus    `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	AdminNotes   *string      `gorm:"type:text" json:"admin_notes,omitempty"` // Alasan penolakan admin
	ReviewedBy   *string      `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Reviewer     *UserProfile `gorm:"foreignKey:ReviewedBy"`
	ReviewedAt   *time.Time   `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time    `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"default:now()" json:"updated_at"`
}