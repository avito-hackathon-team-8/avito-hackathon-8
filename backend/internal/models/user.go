package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Verified  bool      `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (user *User) BeforeCreate(_ *gorm.DB) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	return nil
}
