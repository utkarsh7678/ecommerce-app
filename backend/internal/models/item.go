package models

import (
	"time"
)

type Item struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Status    string    `gorm:"default:'available'" json:"status"`
	Price     float64   `gorm:"not null" json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
