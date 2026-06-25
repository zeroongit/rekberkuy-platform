package domain

import "time"

type VendorCategory string

const (
	VendorGedung      VendorCategory = "GEDUNG"
	VendorKatering    VendorCategory = "KATERING"
	VendorSoundSystem VendorCategory = "SOUND_SYSTEM"
	VendorDekorasi    VendorCategory = "DEKORASI"
	VendorEO          VendorCategory = "EVENT_ORGANIZER" 
)

type VendorCategoryModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	CreatedAt time.Time `gorm:"default:now()"`
}

type EOProfile = VendorProfile

type VendorProfile struct {
	VendorID     string      `gorm:"type:uuid;primaryKey;not null" json:"vendor_id"`
	UserProfile  UserProfile `gorm:"foreignKey:VendorID;constraint:OnDelete:CASCADE"`
	BusinessName string      `gorm:"type:varchar(255);not null" json:"business_name"`
	Category     string      `gorm:"type:varchar(100);not null" json:"category"` // Mapping dari ENUM string
	IsVerified   bool        `gorm:"type:boolean;default:false" json:"is_verified"`
	CreatedAt    time.Time   `gorm:"default:now()" json:"created_at"`
}

type EventVendorAllocation struct {
	ID               string            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID    string            `gorm:"type:uuid;not null;index" json:"transaction_id"`
	EventTx          TransactionEvents `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	VendorID         string            `gorm:"type:uuid;not null" json:"vendor_id"`
	Vendor           VendorProfile     `gorm:"foreignKey:VendorID"`
	AllocatedAmount  int64             `gorm:"type:bigint;not null" json:"allocated_amount"`
	ActualPaidAmount int64             `gorm:"type:bigint;default:0" json:"actual_paid_amount"`
	Status           string            `gorm:"type:varchar(50);default:'PLEDGED'" json:"status"` 
	CreatedAt        time.Time         `gorm:"default:now()" json:"created_at"`
}

type EventOfficialDetails struct {
	TransactionID string            `gorm:"type:uuid;primaryKey;not null" json:"transaction_id"`
	EventTx       TransactionEvents `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	OrganizerID   string            `gorm:"type:uuid;not null" json:"organizer_id"`
	Organizer     VendorProfile     `gorm:"foreignKey:OrganizerID"`
	ManagementFee int64             `gorm:"type:bigint;not null" json:"management_fee"` 
	ApprovedAt    time.Time         `gorm:"default:now()" json:"approved_at"`
}