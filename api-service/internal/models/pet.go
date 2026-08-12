package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Pet struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	Name      string    `gorm:"type:varchar(35);not null;default:''"`
	Level     int       `gorm:"not null;default:1"`
	Leaves    int64     `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (pet *Pet) BeforeCreate(_ *gorm.DB) error {
	if pet.ID == uuid.Nil {
		pet.ID = uuid.New()
	}

	return nil
}
