package domain

import "time"

type VendorCategory string

const (
	VendorGedung      VendorCategory = "GEDUNG"
	VendorKatering    VendorCategory = "KATERING"
	VendorSoundSystem VendorCategory = "SOUND_SYSTEM"
	VendorDekorasi    VendorCategory = "DEKORASI"
	
	// Tambahan Kunci untuk EO Besar / Resmi
	VendorEO          VendorCategory = "EVENT_ORGANIZER" 
)

// EOProfile adalah ALIAS dari VendorProfile agar kode di Usecase 
// jauh lebih manusiawi saat dibaca (e.g., var organizer domain.EOProfile)
type EOProfile = VendorProfile

// VendorProfile menampung informasi bisnis dari vendor resmi maupun EO Resmi
type VendorProfile struct {
	VendorID     string         `json:"vendor_id"`
	BusinessName string         `json:"business_name"`
	Category     VendorCategory `json:"category"`
	IsVerified   bool           `json:"is_verified"`
	CreatedAt    time.Time      `json:"created_at"`
}

// EventVendorAllocation mencatat porsi dana escrow yang dialokasikan ke vendor / EO Besar
type EventVendorAllocation struct {
	ID               string    `json:"id"`
	TransactionID    string    `json:"transaction_id"`
	VendorID         string    `json:"vendor_id"`
	AllocatedAmount  int64     `json:"allocated_amount"`
	ActualPaidAmount int64     `json:"actual_paid_amount"`
	Status           string    `json:"status"` // PLEDGED, PARTIALLY_PAID, FULLY_PAID, DISPUTED
	CreatedAt        time.Time `json:"created_at"`
}

// EventOfficialDetails mencatat kontrak resmi untuk EO Besar (Kategori Pengeluaran > 10 Juta)
// Ini menampung Management Fee yang sudah disepakati di awal (bukan hasil ngembat surplus)
type EventOfficialDetails struct {
	TransactionID string    `json:"transaction_id"`
	OrganizerID   string    `json:"organizer_id"`   // Relasi ke VendorProfile (Category: VendorEO)
	ManagementFee int64     `json:"management_fee"` // Fee profesional tetap EO Resmi
	ApprovedAt    time.Time `json:"approved_at"`
}