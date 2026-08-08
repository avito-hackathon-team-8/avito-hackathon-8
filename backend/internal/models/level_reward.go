package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LevelRewardStatus string

const (
	LevelRewardStatusLocked   LevelRewardStatus = "LOCKED"
	LevelRewardStatusUnopened LevelRewardStatus = "UNOPENED"
	LevelRewardStatusClaimed  LevelRewardStatus = "CLAIMED"
	LevelRewardStatusFrozen   LevelRewardStatus = "FROZEN"
)

type LevelReward struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;uniqueIndex:idx_level_rewards_user_id_id,priority:2"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_level_rewards_user_level,priority:1;uniqueIndex:idx_level_rewards_user_id_id,priority:1"`
	User           User           `gorm:"constraint:OnDelete:CASCADE"`
	Level          int            `gorm:"type:smallint;not null;check:level BETWEEN 1 AND 10;uniqueIndex:idx_level_rewards_user_level,priority:2"`
	Title          string         `gorm:"type:varchar(128);not null"`
	Description    string         `gorm:"type:text;not null"`
	Category       RewardCategory `gorm:"type:varchar(32);not null;index"`
	ClaimExpiresAt *time.Time     `gorm:"index"`
	ClaimedAt      *time.Time     `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (reward *LevelReward) BeforeCreate(_ *gorm.DB) error {
	if reward.ID == uuid.Nil {
		reward.ID = uuid.New()
	}

	return nil
}
