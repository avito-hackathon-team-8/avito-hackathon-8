package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RewardCategory string

const (
	RewardCategoryAvitoBonus        RewardCategory = "AVITO_BONUS"
	RewardCategoryFreeDelivery      RewardCategory = "FREE_DELIVERY"
	RewardCategoryFreePromotion     RewardCategory = "FREE_PROMOTION"
	RewardCategoryPromotionDiscount RewardCategory = "PROMOTION_DISCOUNT"
	RewardCategoryDeliveryDiscount  RewardCategory = "DELIVERY_DISCOUNT"
)

type RewardSource string

const (
	RewardSourceLevel       RewardSource = "LEVEL"
	RewardSourceChest       RewardSource = "CHEST"
	RewardSourceLeaderboard RewardSource = "LEADERBOARD"
)

type Reward struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_rewards_active,priority:1"`
	User       User           `gorm:"constraint:OnDelete:CASCADE"`
	Title      string         `gorm:"not null"`
	Category   RewardCategory `gorm:"type:varchar(32);not null;index"`
	Source     RewardSource   `gorm:"type:varchar(32);not null"`
	ExpiresAt  time.Time      `gorm:"not null;index:idx_rewards_active,priority:2"`
	RedeemedAt *time.Time     `gorm:"index:idx_rewards_active,priority:3"`
	CreatedAt  time.Time      `gorm:"not null"`
}

func (reward *Reward) BeforeCreate(_ *gorm.DB) error {
	if reward.ID == uuid.Nil {
		reward.ID = uuid.New()
	}

	return nil
}
