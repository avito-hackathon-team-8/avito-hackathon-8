package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TaskLocked     = "LOCKED"
	TaskInProgress = "IN_PROGRESS"
	TaskCompleted  = "COMPLETED"
	TaskClaimed    = "CLAIMED"
	TaskExpired    = "EXPIRED"
)

func IsKnownTaskType(taskType string) bool {
	switch taskType {
	case "VIEW_LISTINGS", "ADD_TO_FAVORITES", "PUBLISH_LISTING", "BOOST_LISTING", "LEAVE_REVIEW", "COMPLETE_DEAL", "ORDER_WITH_DELIVERY":
		return true
	default:
		return false
	}
}

func IsKnownRewardCategory(category string) bool {
	switch category {
	case "AVITO_BONUS", "FREE_DELIVERY", "FREE_PROMOTION", "PROMOTION_DISCOUNT", "DELIVERY_DISCOUNT":
		return true
	default:
		return false
	}
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Verified  bool      `gorm:"not null;default:false"`
	Interests string    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time
}

type DailyTaskDefinition struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code        string    `gorm:"type:varchar(64);not null;uniqueIndex"`
	Title       string    `gorm:"not null"`
	Slot        int       `gorm:"not null;index"`
	Type        string    `gorm:"type:varchar(64);not null;index"`
	TargetCount int       `gorm:"not null"`
	Reward      int       `gorm:"not null"`
	UnlockLevel int       `gorm:"not null"`
	Categories  string    `gorm:"type:jsonb;not null;default:'[]'"`
	Active      bool      `gorm:"not null;default:true"`
}

func (task *DailyTaskDefinition) BeforeCreate(_ *gorm.DB) error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	return nil
}

type UserDailyTask struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_user_daily_task,priority:1"`
	TaskDefinitionID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_user_daily_task,priority:3"`
	Day              time.Time  `gorm:"type:date;not null;uniqueIndex:idx_user_daily_task,priority:2"`
	Status           string     `gorm:"type:varchar(16);not null"`
	CurrentCount     int        `gorm:"not null;default:0"`
	CompletedAt      *time.Time `gorm:"index"`
	ClaimedAt        *time.Time `gorm:"index"`
	ExpiresAt        time.Time  `gorm:"not null;index"`
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
}

func (task *UserDailyTask) BeforeCreate(_ *gorm.DB) error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	return nil
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

type LeafTransaction struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_leaf_period,priority:1"`
	Amount       int64     `gorm:"not null"`
	Reason       string    `gorm:"type:varchar(32);not null;index"`
	OperationKey string    `gorm:"type:varchar(160);not null;uniqueIndex"`
	OccurredAt   time.Time `gorm:"not null;index:idx_leaf_period,priority:2"`
	CreatedAt    time.Time `gorm:"not null"`
}

type Reward struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_rewards_active,priority:1;uniqueIndex:idx_reward_source_ref,priority:1"`
	Title           string     `gorm:"not null"`
	Category        string     `gorm:"type:varchar(32);not null;index"`
	Source          string     `gorm:"type:varchar(32);not null;uniqueIndex:idx_reward_source_ref,priority:2"`
	SourceReference *string    `gorm:"type:varchar(128);uniqueIndex:idx_reward_source_ref,priority:3"`
	ExpiresAt       time.Time  `gorm:"not null;index:idx_rewards_active,priority:2"`
	RedeemedAt      *time.Time `gorm:"index:idx_rewards_active,priority:3"`
	CreatedAt       time.Time  `gorm:"not null"`
}

func (reward *Reward) BeforeCreate(_ *gorm.DB) error {
	if reward.ID == uuid.Nil {
		reward.ID = uuid.New()
	}

	return nil
}

type JobRun struct {
	JobName string    `gorm:"type:varchar(128);primaryKey"`
	RunDay  time.Time `gorm:"type:date;primaryKey"`
	RanAt   time.Time `gorm:"not null"`
}
