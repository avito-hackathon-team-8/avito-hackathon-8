package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserLogin struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_activity_user_date,priority:1"`
	User         User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ActivityDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_user_activity_user_date,priority:2"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (activity *UserLogin) BeforeCreate(_ *gorm.DB) error {
	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}

	return nil
}
