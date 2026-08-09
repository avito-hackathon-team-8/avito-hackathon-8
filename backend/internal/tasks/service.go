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
	ErrInvalidTaskType      = errors.New("invalid task type")
	ErrTaskNotFound         = errors.New("task not found")
	ErrTaskLocked           = errors.New("task is locked")
	ErrTaskNotCompleted     = errors.New("task is not completed")
	ErrRewardAlreadyClaimed = errors.New("task reward already claimed")
	ErrTasksNotReady        = errors.New("daily tasks are not ready")
)

const (
	TotalDailyTasks        = 4
	DemoCompletedTaskCount = 2
)

type AssignmentEnsurer interface {
	EnsureDailyTasks(context.Context, uuid.UUID) error
}

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
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

type Event struct {
	TaskID uuid.UUID
	Type   models.TaskType
	Count  int
}

type Service struct {
	db          *gorm.DB
	dailyReport DailyReportNotifier
	now         func() time.Time
	ensurer     AssignmentEnsurer
}

func NewService(db *gorm.DB, dailyReport DailyReportNotifier, ensurer ...AssignmentEnsurer) *Service {
	service := &Service{db: db, dailyReport: dailyReport, now: time.Now}

	if len(ensurer) > 0 {
		service.ensurer = ensurer[0]
	}

	return service
}

func (service *Service) List(ctx context.Context, userID uuid.UUID, userLevel int) ([]DailyTask, error) {
	rows, err := service.rows(ctx, userID, service.today())

	if err != nil {
		return nil, err
	}

	if len(rows) < TotalDailyTasks && service.ensurer != nil {
		if err := service.ensurer.EnsureDailyTasks(ctx, userID); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrTasksNotReady, err)
		}

		rows, err = service.rows(ctx, userID, service.today())

		if err != nil {
			return nil, err
		}
	}

	if len(rows) != TotalDailyTasks {
		return nil, ErrTasksNotReady
	}

	result := make([]DailyTask, 0, len(rows))

	for _, row := range rows {
		status := row.Status

		if userLevel < row.UnlockLevel {
			status = models.LockedTaskStatus
		} else if status == models.LockedTaskStatus {
			status = models.InProgressTaskStatus
		}

		result = append(result, DailyTask{
			ID: row.AssignmentID, Slot: row.Slot, Type: row.Type, Description: row.Description,
			CurrentCount: min(row.CurrentCount, row.TargetCount), TargetCount: row.TargetCount,
			RewardLeaves: row.Reward, RequiredLevel: row.UnlockLevel, Status: status,
		})
	}

	return result, nil
}

func (service *Service) AutoCompleteFirstTasks(ctx context.Context, userID uuid.UUID) error {
	today := service.today()
	rows, err := service.rows(ctx, userID, today)

	if err != nil {
		return err
	}

	if len(rows) != TotalDailyTasks {
		return ErrTasksNotReady
	}

	return service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, row := range rows[:DemoCompletedTaskCount] {
			if err := service.autoCompleteTaskTx(tx, userID, today, row); err != nil {
				return err
			}
		}

		return nil
	})
}

func (service *Service) autoCompleteTaskTx(tx *gorm.DB, userID uuid.UUID, day time.Time, row taskRow) error {
	if row.Status == models.CompletedTaskStatus || row.Status == models.ClaimedTaskStatus {
		return nil
	}

	now := service.now().UTC()
	result := tx.
		Model(&models.UserDailyTask{}).
		Where("id = ? AND user_id = ? AND day = ? AND claimed_at IS NULL AND completed_at IS NULL", row.AssignmentID, userID, utcDate(day)).
		Updates(map[string]any{
			"current_count": row.TargetCount,
			"status":        models.CompletedTaskStatus,
			"completed_at":  now,
		})

	if result.Error != nil {
		return fmt.Errorf("complete first daily task: %w", result.Error)
	}

	return nil
}

func (service *Service) Progress(ctx context.Context, userID uuid.UUID, userLevel int) (DailyProgress, error) {
	dailyTasks, err := service.List(ctx, userID, userLevel)

	if err != nil {
		return DailyProgress{}, err
	}

	completed := 0

	for _, task := range dailyTasks {
		if task.Status == models.CompletedTaskStatus || task.Status == models.ClaimedTaskStatus {
			completed++
		}
	}

	return DailyProgress{CompletedCount: completed, TotalCount: TotalDailyTasks}, nil
}

func (service *Service) RecordEvents(ctx context.Context, userID uuid.UUID, events []Event, userLevel int) error {
	return service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return service.RecordEventsTx(tx, userID, events, userLevel, service.today())
	})
}

func (service *Service) RecordEventsTx(tx *gorm.DB, userID uuid.UUID, events []Event, userLevel int, day time.Time) error {
	for _, event := range events {
		if event.TaskID == uuid.Nil && !models.IsKnownTaskType(event.Type) {
			return ErrInvalidTaskType
		}

		var assignment models.UserDailyTask

		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND day = ?", userID, utcDate(day))

		if event.TaskID != uuid.Nil {
			query = query.Where("id = ?", event.TaskID)
		} else {
			query = query.Where("task_definition_id IN (SELECT id FROM daily_task_definitions WHERE type = ?)", event.Type)
		}

		if err := query.First(&assignment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTaskNotFound
			}

			return err
		}

		var definition models.DailyTaskDefinition

		if err := tx.First(&definition, "id = ?", assignment.TaskDefinitionID).Error; err != nil {
			return err
		}

		if userLevel < definition.UnlockLevel {
			return ErrTaskLocked
		}

		if assignment.ClaimedAt != nil || assignment.CompletedAt != nil {
			continue
		}

		count := event.Count

		if count < 1 {
			count = 1
		}

		remaining := definition.TargetCount - assignment.CurrentCount

		if remaining <= 0 || count >= remaining {
			assignment.CurrentCount = definition.TargetCount
		} else {
			assignment.CurrentCount += count
		}

		if assignment.CurrentCount >= definition.TargetCount {
			now := service.now().UTC()

			assignment.CompletedAt = &now
			assignment.Status = models.CompletedTaskStatus
		}

		if err := tx.Save(&assignment).Error; err != nil {
			return err
		}
	}

	return nil
}

func (service *Service) ClaimWithReward(ctx context.Context, userID, taskID uuid.UUID, userLevel int, applyReward func(*gorm.DB, int) error) (ClaimResult, error) {
	var result ClaimResult

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var assignment models.UserDailyTask

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ? AND day = ?", taskID, userID, service.today()).First(&assignment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTaskNotFound
			}

			return err
		}

		var definition models.DailyTaskDefinition

		if err := tx.First(&definition, "id = ?", assignment.TaskDefinitionID).Error; err != nil {
			return err
		}

		if userLevel < definition.UnlockLevel {
			return ErrTaskLocked
		}

		if assignment.ClaimedAt != nil {
			return ErrRewardAlreadyClaimed
		}

		if assignment.CompletedAt == nil || assignment.CurrentCount < definition.TargetCount {
			return ErrTaskNotCompleted
		}

		now := service.now().UTC()

		assignment.ClaimedAt = &now
		assignment.Status = models.ClaimedTaskStatus

		if err := tx.Save(&assignment).Error; err != nil {
			return err
		}

		result = ClaimResult{TaskID: taskID, RewardLeaves: definition.Reward, Status: models.ClaimedTaskStatus}

		if applyReward != nil {
			return applyReward(tx, definition.Reward)
		}

		return nil
	})

	if err != nil {
		return ClaimResult{}, err
	}

	service.dailyReport.Notify(userID)

	return result, nil
}

type taskRow struct {
	AssignmentID uuid.UUID `gorm:"column:assignment_id"`
	Slot         int
	Type         models.TaskType
	Description  string
	TargetCount  int
	Reward       int
	UnlockLevel  int
	Status       models.TaskStatus
	CurrentCount int
}

func (service *Service) rows(ctx context.Context, userID uuid.UUID, day time.Time) ([]taskRow, error) {
	var rows []taskRow

	err := service.db.WithContext(ctx).Table("user_daily_tasks AS assignments").Select(`
		assignments.id AS assignment_id, definitions.slot, definitions.type,
		definitions.title AS description, definitions.target_count, definitions.reward,
		definitions.unlock_level, assignments.status, assignments.current_count`).
		Joins("JOIN daily_task_definitions AS definitions ON definitions.id = assignments.task_definition_id").
		Where("assignments.user_id = ? AND assignments.day = ?", userID, utcDate(day)).
		Order("definitions.slot ASC").Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("list assigned tasks: %w", err)
	}

	return rows, nil
}

func (service *Service) today() time.Time {
	return utcDate(service.now())
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()

	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
