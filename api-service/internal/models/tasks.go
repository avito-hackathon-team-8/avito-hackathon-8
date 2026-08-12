package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string
type TaskType string

const (
	LockedTaskStatus     TaskStatus = "LOCKED"
	ClaimedTaskStatus    TaskStatus = "CLAIMED"
	InProgressTaskStatus TaskStatus = "IN_PROGRESS"
	CompletedTaskStatus  TaskStatus = "COMPLETED"
)

const (
	ViewListingsTaskType      TaskType = "VIEW_LISTINGS"
	AddToFavoritesTaskType    TaskType = "ADD_TO_FAVORITES"
	PublishListingTaskType    TaskType = "PUBLISH_LISTING"
	BoostListingTaskType      TaskType = "BOOST_LISTING"
	LeaveReviewTaskType       TaskType = "LEAVE_REVIEW"
	CompleteDealTaskType      TaskType = "COMPLETE_DEAL"
	OrderWithDeliveryTaskType TaskType = "ORDER_WITH_DELIVERY"
)

func IsKnownTaskType(taskType TaskType) bool {
	switch taskType {
	case ViewListingsTaskType, AddToFavoritesTaskType, PublishListingTaskType,
		BoostListingTaskType, LeaveReviewTaskType, CompleteDealTaskType, OrderWithDeliveryTaskType:
		return true
	default:
		return false
	}
}

type DailyTaskDefinition struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code        string    `gorm:"type:varchar(64);not null;uniqueIndex"`
	Title       string    `gorm:"not null"`
	Slot        int       `gorm:"not null;index"`
	Type        TaskType  `gorm:"type:varchar(64);not null;index"`
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
	Status           TaskStatus `gorm:"type:varchar(16);not null"`
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
