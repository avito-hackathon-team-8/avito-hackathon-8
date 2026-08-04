package database

import (
	"github.com/mister-cpp/avito-hackathon-8/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{}, &models.OTP{}); err != nil {
		return nil, err
	}

	return db, nil
}
