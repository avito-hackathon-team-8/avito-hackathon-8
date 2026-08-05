package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

type TaskType string

const (
	LockedTaskStatus     TaskStatus = "locked"
	ClaimedTaskStatus    TaskStatus = "claimed"
	InProgressTaskStatus TaskStatus = "in_progress"
	CompletedTaskStatus  TaskStatus = "completed"
)

const (
	OpenNotificationTaskType       TaskType = "open_notification"
	AddToFavoritesTaskType         TaskType = "add_to_favorites"
	PublishListingTaskType         TaskType = "publish_listing"
	BoostListingTaskType           TaskType = "boost_listing"
	LeaveReviewTaskType            TaskType = "leave_review"
	CompleteDealTaskType           TaskType = "complete_deal"
	OrderWithAvitoDeliveryTaskType TaskType = "order_with_avito_delivery"
)

type Task struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey"`
	Date          time.Time          `gorm:"not null"`
	Slot          int                `gorm:"not null"`
	Type          TaskType           `gorm:"type:varchar(50);not null"`
	Description   string             `gorm:"type:text;not null"`
	TargetCount   int                `gorm:"not null"`
	RewardLeaves  int                `gorm:"not null"`
	RequiredLevel int                `gorm:"not null"`
	UserProgress  []UserTaskProgress `gorm:"foreignKey:TaskID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (task *Task) BeforeCreate() error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	return nil
}

func (task Task) StatusFor()
