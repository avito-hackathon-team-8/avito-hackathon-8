package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTP struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	User           User      `gorm:"constraint:OnDelete:CASCADE"`
	CodeHash       string    `gorm:"not null"`
	FailedAttempts int       `gorm:"not null;default:0"`
	ExpiresAt      time.Time `gorm:"not null;index"`
	CreatedAt      time.Time
}

func (otp *OTP) BeforeCreate(_ *gorm.DB) error {
	if otp.ID == uuid.Nil {
		otp.ID = uuid.New()
	}

	return nil
}
