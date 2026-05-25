package domain

import (
	"time"
)

// UserRole mencerminkan ENUM user_role di Supabase
type UserRole string

const (
	RoleMember           UserRole = "MEMBER"
	RoleVerifiedMerchant UserRole = "VERIFIED_MERCHANT"
	RoleVerifiedVendor   UserRole = "VERIFIED_VENDOR"
	RoleAdmin            UserRole = "ADMIN"
	RoleMediator         UserRole = "MEDIATOR"
)

// UserProfile mewakili tabel 'user_profiles'
type UserProfile struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	FullName      string    `json:"full_name"`
	Role          UserRole  `json:"role"`
	PhoneNumber   *string   `json:"phone_number,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CRMLoyalty mewakili tabel 'crm_loyalty' (Tempat hitung koin Shopee-style kita)
type CRMLoyalty struct {
	UserID                   string    `json:"user_id"`
	TotalPoints              int64     `json:"total_points"`
	CurrentTier              string    `json:"current_tier"` // NEWBIE, SILVER, GOLD
	TotalSpentFiat           int64     `json:"total_spent_fiat"`
	UpdatedAt                time.Time `json:"updated_at"`
	
	// Tambahan Field Baru untuk Mengatasi Error Kompiler Go
	TotalCompletedServices   int       `json:"total_completed_services"`   // Jumlah proyek jasa sukses
	TotalCompletedEvents    int       `json:"total_completed_events"`
	ConsecutiveFailedMonths  int       `json:"consecutive_failed_months"`  // Sisa nyawa (maks 3)
	CurrentMonthGmv          int64     `json:"current_month_gmv"`          // Penjual bulanan berjalan
	MaxItemPriceSold         int64     `json:"max_item_price_sold"`        // Detektor klaster harga barang
	TierEvaluationStartedAt  time.Time `json:"tier_evaluation_started_at"`
	LastMonthEvaluatedAt     time.Time `json:"last_month_evaluated_at"`
}