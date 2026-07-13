package domain

import "time"

// ============================================================================
// 🛍️ KATEGORI BARANG (GOODS)
// ============================================================================

type GoodsCategory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Slug      string    `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type GoodsSubCategory struct {
	ID         uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID uint64        `gorm:"not null" json:"category_id"`
	Category   GoodsCategory `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"-"`
	Name       string        `gorm:"type:varchar(100);not null" json:"name"`
	Slug       string        `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt  time.Time     `json:"created_at"`
}

type GoodsSubSubCategory struct {
	ID            uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	SubCategoryID uint64           `gorm:"not null" json:"sub_category_id"`
	SubCategory   GoodsSubCategory `gorm:"foreignKey:SubCategoryID;constraint:OnDelete:CASCADE" json:"-"`
	Name          string           `gorm:"type:varchar(100);not null" json:"name"`
	Slug          string           `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt     time.Time        `json:"created_at"`
}

// ============================================================================
// 💼 KATEGORI JASA (SERVICES)
// ============================================================================

type ServiceCategory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Slug      string    `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type ServiceSubCategory struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID uint64          `gorm:"not null" json:"category_id"`
	Category   ServiceCategory `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"-"`
	Name       string          `gorm:"type:varchar(100);not null" json:"name"`
	Slug       string          `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ServiceSubSubCategory struct {
	ID            uint64             `gorm:"primaryKey;autoIncrement" json:"id"`
	SubCategoryID uint64             `gorm:"not null" json:"sub_category_id"`
	SubCategory   ServiceSubCategory `gorm:"foreignKey:SubCategoryID;constraint:OnDelete:CASCADE" json:"-"`
	Name          string             `gorm:"type:varchar(100);not null" json:"name"`
	Slug          string             `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt     time.Time          `json:"created_at"`
}

// ============================================================================
// 🎪 KATEGORI ACARA (EVENTS)
// ============================================================================

type EventCategory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Slug      string    `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type EventSubCategory struct {
	ID         uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID uint64        `gorm:"not null" json:"category_id"`
	Category   EventCategory `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"-"`
	Name       string        `gorm:"type:varchar(100);not null" json:"name"`
	Slug       string        `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt  time.Time     `json:"created_at"`
}

type EventSubSubCategory struct {
	ID            uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	SubCategoryID uint64           `gorm:"not null" json:"sub_category_id"`
	SubCategory   EventSubCategory `gorm:"foreignKey:SubCategoryID;constraint:OnDelete:CASCADE" json:"-"`
	Name          string           `gorm:"type:varchar(100);not null" json:"name"`
	Slug          string           `gorm:"type:varchar(100);not null;unique" json:"slug"`
	CreatedAt     time.Time        `json:"created_at"`
}

// ============================================================================
// 🏢 KATEGORI MITRA BISNIS (VENDORS)
// ============================================================================

type VendorCategoryModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	CreatedAt time.Time `gorm:"default:now()"`
}