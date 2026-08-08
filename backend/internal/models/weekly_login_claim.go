package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DayStatus string

const (
	DayStatusClaimed   DayStatus = "CLAIMED"
	DayStatusAvailable DayStatus = "AVAILABLE"
	DayStatusMissed    DayStatus = "MISSED"
	DayStatusFuture    DayStatus = "FUTURE"
)

type WeeklyLoginClaim struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_weekly_login_claim_user_date,priority:1"`
	User         User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ClaimDate    time.Time `gorm:"type:date;not null;uniqueIndex:idx_weekly_login_claim_user_date,priority:2"`
	RewardLeaves int       `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (claim *WeeklyLoginClaim) BeforeCreate(_ *gorm.DB) error {
	if claim.ID == uuid.Nil {
		claim.ID = uuid.New()
	}

	return nil
}
