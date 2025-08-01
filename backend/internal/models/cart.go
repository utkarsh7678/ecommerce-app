package models

import (
	"time"
)

type Cart struct {
	ID        uint       `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	UserID    *uint      `gorm:"default:null" json:"user_id"`
	SessionID string     `gorm:"size:255;default:'';index" json:"-"`
	Status    string     `gorm:"default:'active'" json:"status"`
	Items     []CartItem `gorm:"foreignkey:CartID" json:"items,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type CartItem struct {
	CartID    uint      `gorm:"primaryKey;index" json:"-"`
	ItemID    uint      `gorm:"primaryKey" json:"item_id"`
	Quantity  int       `gorm:"default:1" json:"quantity"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
