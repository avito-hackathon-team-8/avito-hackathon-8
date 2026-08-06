package database

import (
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Pet{},
		&models.OTP{},
		&models.LevelReward{},
		&models.Reward{},
		&models.Task{},
		&models.UserTaskProgress{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
