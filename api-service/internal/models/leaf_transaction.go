package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeafTransactionReason string

const (
	LeafReasonTaskReward    LeafTransactionReason = "TASK_REWARD"
	LeafReasonWeeklyLogin   LeafTransactionReason = "WEEKLY_LOGIN"
	LeafReasonLevelUp       LeafTransactionReason = "LEVEL_UP"
	LeafReasonChestPurchase LeafTransactionReason = "CHEST_PURCHASE"
	LeafReasonShopPurchase  LeafTransactionReason = "SHOP_PURCHASE"
	LeafReasonMVP           LeafTransactionReason = "MVP"
)

type LeafTransaction struct {
	ID           uuid.UUID             `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID             `gorm:"type:uuid;not null;index:idx_leaf_period,priority:1"`
	Amount       int64                 `gorm:"not null"`
	Reason       LeafTransactionReason `gorm:"type:varchar(32);not null;index"`
	OperationKey string                `gorm:"type:varchar(160);not null;uniqueIndex"`
	OccurredAt   time.Time             `gorm:"not null;index:idx_leaf_period,priority:2"`
	CreatedAt    time.Time             `gorm:"not null"`
}

func (transaction *LeafTransaction) BeforeCreate(_ *gorm.DB) error {
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}

	return nil
}
