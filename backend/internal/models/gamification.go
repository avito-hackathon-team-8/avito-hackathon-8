package models

import (
	"time"

	"github.com/google/uuid"
)

type UserGameState struct {
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	PetLevel    int       `gorm:"not null;default:1"`
	LeafBalance int64     `gorm:"not null;default:0"`
	UpdatedAt   time.Time `gorm:"not null"`
}

type LeaderboardEntry struct {
	PeriodStart  time.Time `gorm:"type:date;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Leaves       int64     `gorm:"not null"`
	Rank         int64     `gorm:"not null;index"`
	CalculatedAt time.Time `gorm:"not null"`
}

type LeaderboardSeason struct {
	PeriodStart time.Time  `gorm:"type:date;primaryKey"`
	FinalizedAt *time.Time `gorm:"index"`
}

type JobRun struct {
	JobName string    `gorm:"type:varchar(128);primaryKey"`
	RunDay  time.Time `gorm:"type:date;primaryKey"`
	RanAt   time.Time `gorm:"not null"`
}
