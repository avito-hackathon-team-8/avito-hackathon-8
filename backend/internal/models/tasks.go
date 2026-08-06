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

type Task struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey"`
	Date          time.Time          `gorm:"type:date;not null;uniqueIndex:idx_tasks_date_slot,priority:1"`
	Slot          int                `gorm:"not null;uniqueIndex:idx_tasks_date_slot,priority:2"`
	Type          TaskType           `gorm:"type:varchar(64);not null;index"`
	Description   string             `gorm:"type:text;not null"`
	TargetCount   int                `gorm:"not null"`
	RewardLeaves  int                `gorm:"not null"`
	RequiredLevel int                `gorm:"not null"`
	UserProgress  []UserTaskProgress `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserTaskProgress struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TaskID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_task_progress_user_task,unique"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_task_progress_user_task,unique"`
	CurrentCount int        `gorm:"not null;default:0"`
	CompletedAt  *time.Time `gorm:"index"`
	ClaimedAt    *time.Time `gorm:"index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (task *Task) BeforeCreate(_ *gorm.DB) error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	return nil
}

func (progress *UserTaskProgress) BeforeCreate(_ *gorm.DB) error {
	if progress.ID == uuid.Nil {
		progress.ID = uuid.New()
	}

	return nil
}
