package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaderboardTotal struct {
	PeriodStart time.Time `gorm:"type:date;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	Leaves      int64     `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (LeaderboardTotal) TableName() string { return "leaderboard_totals" }
