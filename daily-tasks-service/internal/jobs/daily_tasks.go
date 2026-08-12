package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrAssignmentUserNotFound = errors.New("verified user not found")

const assignmentBatchSize = 500

type TaskAssigner struct {
	db          *gorm.DB
	definitions []models.DailyTaskDefinition
	selector    TaskSelector
	now         func() time.Time
	observer    WorkObserver
}

type WorkObserver interface {
	AddUsers(job string, count int)
	AddRows(job string, count int)
}

func NewTaskAssigner(db *gorm.DB, definitions []models.DailyTaskDefinition, selector TaskSelector, observers ...WorkObserver) *TaskAssigner {
	if selector == nil {
		selector = SimilaritySelector{}
	}

	var observer WorkObserver

	if len(observers) > 0 {
		observer = observers[0]
	}

	return &TaskAssigner{db: db, definitions: definitions, selector: selector, now: time.Now, observer: observer}
}

func (assigner *TaskAssigner) Seed(ctx context.Context) error {
	return assigner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return assigner.seedTx(tx)
	})
}

func (assigner *TaskAssigner) Ensure(ctx context.Context, userID uuid.UUID) error {
	return assigner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

	var definitions []models.DailyTaskDefinition

	if err := tx.Where("active = ?", true).Order("slot, code").Find(&definitions).Error; err != nil {
		return fmt.Errorf("load task definitions: %w", err)
	}

	var after uuid.UUID

	for batch := 1; ; batch++ {
		var users []models.User

		query := assigner.db.WithContext(tx.Statement.Context).Where("verified = ?", true).Order("id").Limit(assignmentBatchSize)

		if after != uuid.Nil {
			query = query.Where("id > ?", after)
		}

		if err := query.Find(&users).Error; err != nil {
			return fmt.Errorf("load users batch %d: %w", batch, err)
		}

		if len(users) == 0 {
			return nil
		}

		if err := assigner.db.WithContext(tx.Statement.Context).Transaction(func(batchTx *gorm.DB) error { return assigner.ensureBatchTx(batchTx, users, definitions, now) }); err != nil {
			return fmt.Errorf("assign users batch %d: %w", batch, err)
		}

		if assigner.observer != nil {
			assigner.observer.AddUsers("distribute-daily-tasks", len(users))
		}

		after = users[len(users)-1].ID
	}
}

func (assigner *TaskAssigner) ensureBatchTx(tx *gorm.DB, users []models.User, definitions []models.DailyTaskDefinition, now time.Time) error {
	ids := make([]uuid.UUID, 0, len(users))

	for _, user := range users {
		ids = append(ids, user.ID)
	}

	var petRows []struct {
		UserID uuid.UUID
		Level  int
	}

	if err := tx.Table("pets").Select("user_id, level").Where("user_id IN ?", ids).Scan(&petRows).Error; err != nil {
		return fmt.Errorf("load pet levels: %w", err)
	}

	levels := make(map[uuid.UUID]int, len(users))

	for _, user := range users {
		levels[user.ID] = 1
	}

	for _, row := range petRows {
		levels[row.UserID] = row.Level
	}

	day := startOfDay(now.UTC())
	nextDay := day.AddDate(0, 0, 1)

	var occupiedRows []struct {
		UserID uuid.UUID
		Slot   int
	}

	if err := tx.Table("user_daily_tasks AS assignments").Select("assignments.user_id, definitions.slot").Joins("JOIN daily_task_definitions AS definitions ON definitions.id = assignments.task_definition_id").Where("assignments.user_id IN ? AND assignments.day = ?", ids, day).Scan(&occupiedRows).Error; err != nil {
		return fmt.Errorf("load occupied task slots: %w", err)
	}

	occupied := make(map[uuid.UUID]map[int]struct{}, len(users))

	for _, row := range occupiedRows {
		if occupied[row.UserID] == nil {
			occupied[row.UserID] = make(map[int]struct{})
		}

		occupied[row.UserID][row.Slot] = struct{}{}
	}

	assignments := make([]models.UserDailyTask, 0, len(users)*4)

	for _, user := range users {
		slots := occupied[user.ID]

		if slots == nil {
			slots = make(map[int]struct{})
		}
		for _, definition := range assigner.selector.Select(user, definitions, day) {
			if _, exists := slots[definition.Slot]; exists {
				continue
			}

			status := models.TaskInProgress

			if levels[user.ID] < definition.UnlockLevel {
				status = models.TaskLocked
			}

			assignments = append(assignments, models.UserDailyTask{ID: uuid.New(), UserID: user.ID, TaskDefinitionID: definition.ID, Day: day, Status: status, ExpiresAt: nextDay, CreatedAt: now.UTC(), UpdatedAt: now.UTC()})
			slots[definition.Slot] = struct{}{}
		}
	}

	if len(assignments) == 0 {
		return nil
	}

	result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "day"}, {Name: "task_definition_id"}}, DoNothing: true}).CreateInBatches(&assignments, 500)

	if result.Error == nil && assigner.observer != nil {
		assigner.observer.AddRows("distribute-daily-tasks", int(result.RowsAffected))
	}

	return result.Error
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

	petLevel := 1

	var petRow struct {
		Level int
	}

	if err := tx.Table("pets").Select("level").Where("user_id = ?", userID).Take(&petRow).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load pet: %w", err)
		}
	} else {
		petLevel = petRow.Level
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

		if petLevel < definition.UnlockLevel {
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
