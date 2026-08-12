package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaderboardTotal struct {
	PeriodStart time.Time `gorm:"type:date;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	Leaves      int64
	UpdatedAt   time.Time
}

func (LeaderboardTotal) TableName() string { return "leaderboard_totals" }
