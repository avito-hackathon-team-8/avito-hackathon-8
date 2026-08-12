package models

import (
	"time"

	"github.com/google/uuid"
)

type ExternalEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Type        string    `gorm:"type:varchar(64);not null"`
	Count       int       `gorm:"not null"`
	OccurredAt  time.Time `gorm:"not null;index"`
	PayloadHash string    `gorm:"type:char(64);not null"`
	CreatedAt   time.Time `gorm:"not null"`
}
