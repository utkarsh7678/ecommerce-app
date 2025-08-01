package config

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite", "file:ecommerce.db?cache=shared&mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Enable logging
	db.LogMode(true)

	// Set connection pool settings
	sqlDB := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	return db, nil
}

func GetDB() *gorm.DB {
	return DB
}
