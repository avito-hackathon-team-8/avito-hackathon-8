package tasks

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidTaskType      = errors.New("invalid task type")
	ErrTaskNotFound         = errors.New("task not found")
	ErrTaskLocked           = errors.New("task is locked")
	ErrTaskNotCompleted     = errors.New("task is not completed")
	ErrRewardAlreadyClaimed = errors.New("task reward already claimed")
)

const TotalDailyTasks = 4

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

type Event struct {
	Type  models.TaskType
	Count int
}

type Service struct {
	db          *gorm.DB
	now         func() time.Time
	definitions []Definition
}

func NewService(db *gorm.DB, definitions []Definition) *Service {
	return &Service{
		db:          db,
		now:         time.Now,
		definitions: append([]Definition(nil), definitions...),
	}
}

func (service *Service) List(ctx context.Context, userID uuid.UUID, userLevel int) ([]DailyTask, error) {
	date := service.today()

	if err := service.ensureTasks(ctx, date); err != nil {
		return nil, err
	}

	var tasks []models.Task

	if err := service.db.WithContext(ctx).
		Where("date = ?", date).
		Order("slot ASC").
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list daily tasks: %w", err)
	}

	var progress []models.UserTaskProgress

	if err := service.db.WithContext(ctx).
		Where("user_id = ? AND task_id IN (?)", userID, taskIDs(tasks)).
		Find(&progress).Error; err != nil {
		return nil, fmt.Errorf("list daily task progress: %w", err)
	}

	dailyTasks := make([]DailyTask, 0, len(tasks))

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

		dailyTasks = append(dailyTasks, DailyTask{
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

	return dailyTasks, nil
}

func (service *Service) Progress(ctx context.Context, userID uuid.UUID, userLevel int) (DailyProgress, error) {
	tasks, err := service.List(ctx, userID, userLevel)
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

func (service *Service) RecordEvent(
	ctx context.Context,
	userID uuid.UUID,
	taskType models.TaskType,
	userLevel int,
) error {
	return service.RecordEvents(ctx, userID, []Event{{Type: taskType, Count: 1}}, userLevel)
}

func (service *Service) RecordEvents(
	ctx context.Context,
	userID uuid.UUID,
	events []Event,
	userLevel int,
) error {
	if len(events) == 0 {
		return nil
	}

	date := service.today()

	if err := service.ensureTasks(ctx, date); err != nil {
		return err
	}

	return service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			if event.Count < 1 {
				event.Count = 1
			}

			var task models.Task
			err := tx.Where("date = ? AND type = ?", date, event.Type).First(&task).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if !IsKnownTaskType(event.Type) {
					return ErrInvalidTaskType
				}

				return ErrTaskNotFound
			}

			if err != nil {
				return fmt.Errorf("find task: %w", err)
			}

			if userLevel < task.RequiredLevel {
				return ErrTaskLocked
			}

			var progress models.UserTaskProgress
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ? AND task_id = ?", userID, task.ID).
				First(&progress).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				progress = models.UserTaskProgress{
					UserID:       userID,
					TaskID:       task.ID,
					CurrentCount: 0,
				}

				if task.TargetCount <= event.Count {
					progress.CurrentCount = task.TargetCount
					now := service.now().UTC()
					progress.CompletedAt = &now
				} else {
					progress.CurrentCount = event.Count
				}

				err = tx.Create(&progress).Error

				if err != nil {
					return err
				}

				continue
			}

			if err != nil {
				return err
			}

			if progress.CompletedAt != nil && progress.CurrentCount >= task.TargetCount {
				continue
			}

			newCount := progress.CurrentCount + event.Count
			if newCount > task.TargetCount {
				newCount = task.TargetCount
			}
			progress.CurrentCount = newCount

			if progress.CurrentCount == task.TargetCount && progress.CompletedAt == nil {
				now := service.now().UTC()
				progress.CompletedAt = &now
			}

			err = tx.Save(&progress).Error

			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (service *Service) Claim(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, userLevel int) (ClaimResult, error) {
	return service.claim(ctx, userID, taskID, userLevel, nil)
}

func (service *Service) ClaimWithReward(
	ctx context.Context,
	userID uuid.UUID,
	taskID uuid.UUID,
	userLevel int,
	applyReward func(*gorm.DB, int) error,
) (ClaimResult, error) {
	return service.claim(ctx, userID, taskID, userLevel, applyReward)
}

func (service *Service) claim(
	ctx context.Context,
	userID uuid.UUID,
	taskID uuid.UUID,
	userLevel int,
	applyReward func(*gorm.DB, int) error,
) (ClaimResult, error) {
	date := service.today()

	if err := service.ensureTasks(ctx, date); err != nil {
		return ClaimResult{}, err
	}

	var claimResult ClaimResult

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		now := service.now().UTC()
		err = tx.Model(&progress).Update("claimed_at", now).Error

		if err != nil {
			return err
		}

		claimResult = ClaimResult{
			TaskID:       task.ID,
			RewardLeaves: task.RewardLeaves,
			Status:       models.ClaimedTaskStatus,
		}

		if applyReward != nil {
			err := applyReward(tx, claimResult.RewardLeaves)

			if err != nil {
				return fmt.Errorf("apply task reward: %w", err)
			}
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

	return claimResult, nil
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
	definitions := service.definitionsForDate(date)
	tasks := make([]models.Task, 0, len(definitions))

	for _, definition := range definitions {
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

	err := service.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "date"}, {Name: "slot"}},
			DoNothing: true,
		}).
		Create(&tasks).Error

	if err != nil {
		return fmt.Errorf("ensure daily tasks: %w", err)
	}

	return nil
}

func (service *Service) definitionsForDate(date time.Time) []Definition {
	definitions := make([]Definition, 0, TotalDailyTasks)

	for _, slot := range dailyTaskSlots {
		candidates := make([]Definition, 0)
		for _, definition := range service.definitions {
			if definition.Slot == slot {
				candidates = append(candidates, definition)
			}
		}

		definitions = append(definitions, candidates[definitionIndex(date, slot, len(candidates))])
	}

	return definitions
}

func definitionIndex(date time.Time, slot, candidatesCount int) int {
	source := fmt.Sprintf("%s:%d", date.UTC().Format(time.DateOnly), slot)
	digest := sha256.Sum256([]byte(source))
	value := binary.BigEndian.Uint64(digest[:8])

	return int(value % uint64(candidatesCount))
}

func (service *Service) today() time.Time {
	now := service.now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func taskIDs(tasks []models.Task) []uuid.UUID {
	taskIDs := make([]uuid.UUID, 0, len(tasks))

	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}

	return taskIDs
}
