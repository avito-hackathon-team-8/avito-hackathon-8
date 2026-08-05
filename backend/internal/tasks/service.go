package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTaskNotFound         = errors.New("task not found")
	ErrTaskLocked           = errors.New("task is locked")
	ErrTaskNotCompleted     = errors.New("task is not completed")
	ErrRewardAlreadyClaimed = errors.New("task reward already claimed")
)

const TotalDailyTasks = 4

type Definition struct {
	Slot          int
	Type          models.TaskType
	Description   string
	TargetCount   int
	RewardLeaves  int
	RequiredLevel int
}

var dailyTaskDefinitions = []Definition{
	{
		Slot:          1,
		Type:          models.OpenNotificationsTaskType,
		Description:   "Открыть 5 уведомлений",
		TargetCount:   5,
		RewardLeaves:  45,
		RequiredLevel: 1,
	},
	{
		Slot:          2,
		Type:          models.AddToFavoritesTaskType,
		Description:   "Добавить 3 объявления в избранное",
		TargetCount:   3,
		RewardLeaves:  45,
		RequiredLevel: 1,
	},
	{
		Slot:          3,
		Type:          models.PublishListingTaskType,
		Description:   "Опубликовать объявление",
		TargetCount:   1,
		RewardLeaves:  50,
		RequiredLevel: 5,
	},
	{
		Slot:          4,
		Type:          models.CompleteDealTaskType,
		Description:   "Совершить покупку или продажу",
		TargetCount:   1,
		RewardLeaves:  60,
		RequiredLevel: 10,
	},
}

type DailyTask struct {
	ID            uuid.UUID
	Slot          int
	Type          models.TaskType
	Description   string
	CurrentCount  int
	TargetCount   int
	RewardLeaves  int
	RequiredLevel int
	Status        models.TaskStatus
}

type DailyProgress struct {
	CompletedCount int
	TotalCount     int
}

type ClaimResult struct {
	TaskID       uuid.UUID
	RewardLeaves int
	Status       models.TaskStatus
}

type Service struct {
	db  *gorm.DB
	now func() time.Time
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, now: time.Now}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, userLevel int) ([]DailyTask, error) {
	date := s.today()

	if err := s.ensureTasks(ctx, date); err != nil {
		return nil, err
	}

	var tasks []models.Task
	if err := s.db.WithContext(ctx).
		Where("date = ?", date).
		Order("slot ASC").
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list daily tasks: %w", err)
	}

	var progress []models.UserTaskProgress
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND task_id IN (?)", userID, taskIDs(tasks)).
		Find(&progress).Error; err != nil {
		return nil, fmt.Errorf("list daily task progress: %w", err)
	}

	result := make([]DailyTask, 0, len(tasks))

	for _, task := range tasks {
		var taskProgress *models.UserTaskProgress
		currentCount := 0

		for i := range progress {
			if progress[i].TaskID == task.ID {
				taskProgress = &progress[i]
				currentCount = min(progress[i].CurrentCount, task.TargetCount)
				break
			}
		}

		result = append(result, DailyTask{
			ID:            task.ID,
			Slot:          task.Slot,
			Type:          task.Type,
			Description:   task.Description,
			CurrentCount:  currentCount,
			TargetCount:   task.TargetCount,
			RewardLeaves:  task.RewardLeaves,
			RequiredLevel: task.RequiredLevel,
			Status:        StatusFor(task, taskProgress, userLevel),
		})
	}

	return result, nil
}

func (s *Service) Progress(ctx context.Context, userID uuid.UUID, userLevel int) (DailyProgress, error) {
	tasks, err := s.List(ctx, userID, userLevel)

	if err != nil {
		return DailyProgress{}, err
	}

	completedCount := 0

	for _, task := range tasks {
		if task.Status == models.CompletedTaskStatus || task.Status == models.ClaimedTaskStatus {
			completedCount++
		}
	}

	return DailyProgress{
		CompletedCount: completedCount,
		TotalCount:     TotalDailyTasks,
	}, nil
}

func (s *Service) RecordEvent(
	ctx context.Context,
	userID uuid.UUID,
	taskType models.TaskType,
	userLevel int,
) error {
	date := s.today()

	if err := s.ensureTasks(ctx, date); err != nil {
		return err
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task models.Task

		if err := tx.Where("date = ? AND type = ?", date, taskType).First(&task).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		} else if err != nil {
			return fmt.Errorf("find task: %w", err)
		}

		if userLevel < task.RequiredLevel {
			return ErrTaskLocked
		}

		var progress models.UserTaskProgress

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND task_id = ?", userID, task.ID).
			First(&progress).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			progress = models.UserTaskProgress{
				UserID:       userID,
				TaskID:       task.ID,
				CurrentCount: 1,
			}

			if task.TargetCount <= 1 {
				now := s.now().UTC()
				progress.CurrentCount = task.TargetCount
				progress.CompletedAt = &now
			}

			return tx.Create(&progress).Error
		}

		if err != nil {
			return err
		}

		if progress.CurrentCount < task.TargetCount {
			progress.CurrentCount++
		}

		if progress.CurrentCount == task.TargetCount && progress.CompletedAt == nil {
			now := s.now().UTC()
			progress.CompletedAt = &now
		}

		return tx.Save(&progress).Error
	})

	if errors.Is(err, ErrTaskNotFound) || errors.Is(err, ErrTaskLocked) {
		return err
	}

	if err != nil {
		return fmt.Errorf("record task event: %w", err)
	}

	return nil
}

func (s *Service) Claim(
	ctx context.Context,
	userID uuid.UUID,
	taskID uuid.UUID,
	userLevel int,
) (ClaimResult, error) {
	date := s.today()

	if err := s.ensureTasks(ctx, date); err != nil {
		return ClaimResult{}, err
	}

	var result ClaimResult

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task models.Task

		err := tx.Where("id = ? AND date = ?", taskID, date).First(&task).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}

		if err != nil {
			return err
		}

		if userLevel < task.RequiredLevel {
			return ErrTaskLocked
		}

		var progress models.UserTaskProgress

		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND task_id = ?", userID, task.ID).
			First(&progress).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotCompleted
		}

		if err != nil {
			return err
		}

		if progress.ClaimedAt != nil {
			return ErrRewardAlreadyClaimed
		}

		if progress.CompletedAt == nil || progress.CurrentCount < task.TargetCount {
			return ErrTaskNotCompleted
		}

		now := s.now().UTC()

		if err := tx.Model(&progress).Update("claimed_at", now).Error; err != nil {
			return err
		}

		result = ClaimResult{
			TaskID:       task.ID,
			RewardLeaves: task.RewardLeaves,
			Status:       models.ClaimedTaskStatus,
		}

		return nil
	})

	if errors.Is(err, ErrTaskNotFound) ||
		errors.Is(err, ErrTaskLocked) ||
		errors.Is(err, ErrTaskNotCompleted) ||
		errors.Is(err, ErrRewardAlreadyClaimed) {
		return ClaimResult{}, err
	}

	if err != nil {
		return ClaimResult{}, fmt.Errorf("claim task reward: %w", err)
	}

	return result, nil
}

func StatusFor(task models.Task, progress *models.UserTaskProgress, userLevel int) models.TaskStatus {
	if userLevel < task.RequiredLevel {
		return models.LockedTaskStatus
	}

	if progress == nil {
		return models.InProgressTaskStatus
	}

	if progress.ClaimedAt != nil {
		return models.ClaimedTaskStatus
	}

	if progress.CompletedAt != nil || progress.CurrentCount == task.TargetCount {
		return models.CompletedTaskStatus
	}

	return models.InProgressTaskStatus
}

func (service *Service) ensureTasks(ctx context.Context, date time.Time) error {
	tasks := make([]models.Task, 0, len(dailyTaskDefinitions))

	for _, definition := range dailyTaskDefinitions {
		tasks = append(tasks, models.Task{
			Date:          date,
			Slot:          definition.Slot,
			Type:          definition.Type,
			Description:   definition.Description,
			TargetCount:   definition.TargetCount,
			RewardLeaves:  definition.RewardLeaves,
			RequiredLevel: definition.RequiredLevel,
		})
	}

	if err := service.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "date"}, {Name: "slot"}},
			DoNothing: true,
		}).
		Create(&tasks).Error; err != nil {
		return fmt.Errorf("ensure daily tasks: %w", err)
	}

	return nil
}

func (service *Service) today() time.Time {
	now := service.now().UTC()

	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func taskIDs(tasks []models.Task) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(tasks))

	for _, task := range tasks {
		ids = append(ids, task.ID)
	}

	return ids
}
