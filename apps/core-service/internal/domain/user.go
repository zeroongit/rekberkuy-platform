package domain

import (
	"time"
	"context"
	"github.com/golang-jwt/jwt/v5"
)

type UserRole string

const (
	RoleUser             UserRole = "USER"
	RoleVerifiedMerchant UserRole = "VERIFIED_MERCHANT"
	RoleVerifiedVendor   UserRole = "VERIFIED_VENDOR"
	RoleEventOrganizer   UserRole = "EVENT_ORGANIZER"
	RoleAdmin            UserRole = "ADMIN"
)

type UserProfile struct {
	ID          string    `gorm:"type:uuid;primaryKey;not null" json:"id"`
	Username    string    `gorm:"type:varchar(255);not null;unique" json:"username"`
	FullName    string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Role        UserRole  `gorm:"type:varchar(50);not null;default:'USER'" json:"role"`
	PhoneNumber *string   `gorm:"type:varchar(50)" json:"phone_number,omitempty"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`
}

type CRMLoyalty struct {
	UserID                  string      `gorm:"type:uuid;primaryKey;not null" json:"user_id"`
	UserProfile             UserProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	TotalPoints             int64       `gorm:"type:bigint;not null;default:0" json:"total_points"`
	CurrentTier             string      `gorm:"type:varchar(50);not null;default:'BRONZE'" json:"current_tier"`
	TotalSpentFiat          int64       `gorm:"type:bigint;not null;default:0" json:"total_spent_fiat"`
	Rolling3MonthGMV        int64       `gorm:"type:bigint;not null;default:0" json:"rolling_3_month_gmv"`
	CurrentMonthGmv         int64       `gorm:"type:bigint;not null;default:0" json:"current_month_gmv"`
	MaxItemPriceSold        int64       `gorm:"type:bigint;not null;default:0" json:"max_item_price_sold"`
	TotalCompletedServices  int         `gorm:"type:integer;not null;default:0" json:"total_completed_services"`
	TotalCompletedEvents    int         `gorm:"type:integer;not null;default:0" json:"total_completed_events"`
	ConsecutiveFailedMonths  int         `gorm:"type:integer;not null;default:0" json:"consecutive_failed_months"`
	TierEvaluationStartedAt time.Time   `gorm:"default:now()" json:"tier_evaluation_started_at"`
	LastMonthEvaluatedAt    time.Time   `gorm:"default:now()" json:"last_month_evaluated_at"`
	UpdatedAt               time.Time   `gorm:"default:now()" json:"updated_at"`
}

type JWTCustomClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
	jwt.RegisteredClaims
}

type UserRepository interface {
	CreateProfile(ctx context.Context, user *UserProfile) error
	GetProfileByID(ctx context.Context, id string) (*UserProfile, error)
}