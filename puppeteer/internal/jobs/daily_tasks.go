package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrAssignmentUserNotFound = errors.New("verified user not found")

type TaskAssigner struct {
	db          *gorm.DB
	definitions []models.DailyTaskDefinition
	selector    TaskSelector
	now         func() time.Time
}

func NewTaskAssigner(db *gorm.DB, definitions []models.DailyTaskDefinition, selector TaskSelector) *TaskAssigner {
	if selector == nil {
		selector = SimilaritySelector{}
	}

	return &TaskAssigner{db: db, definitions: definitions, selector: selector, now: time.Now}
}

func (assigner *TaskAssigner) Seed(ctx context.Context) error {
	return assigner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return assigner.seedTx(tx)
	})
}

func (assigner *TaskAssigner) Ensure(ctx context.Context, userID uuid.UUID) error {
	return assigner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := assigner.seedTx(tx); err != nil {
			return err
		}

		return assigner.ensureUserTx(tx, userID, assigner.now().UTC())
	})
}

func (assigner *TaskAssigner) AssignAllTx(tx *gorm.DB, now time.Time) error {
	if err := assigner.seedTx(tx); err != nil {
		return err
	}

	if err := tx.Model(&models.UserDailyTask{}).
		Where("expires_at <= ? AND status IN ?", now.UTC(), []string{models.TaskLocked, models.TaskInProgress, models.TaskCompleted}).
		Update("status", models.TaskExpired).Error; err != nil {
		return fmt.Errorf("expire daily tasks: %w", err)
	}

	var users []models.User

	if err := tx.Where("verified = ?", true).Find(&users).Error; err != nil {
		return fmt.Errorf("load users: %w", err)
	}

	for _, user := range users {
		if err := assigner.ensureUserTx(tx, user.ID, now); err != nil {
			return err
		}
	}

	return nil
}

func (assigner *TaskAssigner) seedTx(tx *gorm.DB) error {
	if err := tx.Model(&models.DailyTaskDefinition{}).Where("active = ?", true).Update("active", false).Error; err != nil {
		return fmt.Errorf("deactivate task definitions: %w", err)
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "type", "slot", "target_count", "reward", "unlock_level", "categories", "active"}),
	}).Create(&assigner.definitions).Error; err != nil {
		return fmt.Errorf("seed task definitions: %w", err)
	}

	return nil
}

func (assigner *TaskAssigner) ensureUserTx(tx *gorm.DB, userID uuid.UUID, now time.Time) error {
	var user models.User

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND verified = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAssignmentUserNotFound
		}

		return err
	}

	if err := tx.Exec(`
		INSERT INTO user_game_states (user_id, pet_level, leaf_balance, updated_at)
		SELECT users.id, COALESCE(pets.level, 1), COALESCE(pets.leaves, 0), ?
		FROM users LEFT JOIN pets ON pets.user_id = users.id
		WHERE users.id = ?
		ON CONFLICT (user_id) DO UPDATE SET pet_level = EXCLUDED.pet_level,
		leaf_balance = EXCLUDED.leaf_balance, updated_at = EXCLUDED.updated_at`, now.UTC(), userID).Error; err != nil {
		return fmt.Errorf("sync game state: %w", err)
	}

	var state models.UserGameState

	if err := tx.First(&state, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("load game state: %w", err)
	}

	var definitions []models.DailyTaskDefinition

	if err := tx.Where("active = ?", true).Order("slot, code").Find(&definitions).Error; err != nil {
		return fmt.Errorf("load task definitions: %w", err)
	}

	day := startOfDay(now.UTC())
	nextDay := day.AddDate(0, 0, 1)

	var occupiedSlots []int

	if err := tx.Table("user_daily_tasks AS assignments").Distinct("definitions.slot").
		Joins("JOIN daily_task_definitions AS definitions ON definitions.id = assignments.task_definition_id").
		Where("assignments.user_id = ? AND assignments.day = ?", userID, day).
		Pluck("definitions.slot", &occupiedSlots).Error; err != nil {
		return fmt.Errorf("load occupied task slots: %w", err)
	}

	occupied := make(map[int]struct{}, len(occupiedSlots))

	for _, slot := range occupiedSlots {
		occupied[slot] = struct{}{}
	}

	for _, definition := range assigner.selector.Select(user, definitions, day) {
		if _, exists := occupied[definition.Slot]; exists {
			continue
		}

		status := models.TaskInProgress

		if state.PetLevel < definition.UnlockLevel {
			status = models.TaskLocked
		}

		assignment := models.UserDailyTask{
			ID: uuid.New(), UserID: user.ID, TaskDefinitionID: definition.ID, Day: day,
			Status: status, ExpiresAt: nextDay, CreatedAt: now.UTC(), UpdatedAt: now.UTC(),
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "day"}, {Name: "task_definition_id"}}, DoNothing: true,
		}).Create(&assignment).Error; err != nil {
			return fmt.Errorf("assign task %s to %s: %w", definition.Code, user.ID, err)
		}

		occupied[definition.Slot] = struct{}{}
	}

	return nil
}

type DistributeDailyTasks struct {
	Assigner *TaskAssigner
}

func (DistributeDailyTasks) Name() string { return "distribute-daily-tasks" }

func (job DistributeDailyTasks) Run(_ context.Context, db *gorm.DB, now time.Time) error {
	if job.Assigner == nil {
		return errors.New("task assigner is not configured")
	}

	return job.Assigner.AssignAllTx(db, now.UTC())
}

func startOfDay(at time.Time) time.Time {
	at = at.UTC()

	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC)
}
