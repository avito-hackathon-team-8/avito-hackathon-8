package daily_report

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type Reward struct {
	ID         uuid.UUID
	Type       models.RewardCategory
	Title      string
	ExpiresAt  time.Time
	ReceivedAt time.Time
}

type Task struct {
	ID            uuid.UUID
	Type          models.TaskType
	Description   string
	RewardLeaves  int
	RewardClaimed bool
	CompletedAt   time.Time
}

type LevelUp struct {
	FromLevel  int
	ToLevel    int
	OccurredAt time.Time
}

type DailyReport struct {
	LeavesEarnedToday int
	Date              string
	Rewards           []Reward
	LevelUp           *LevelUp
	Tasks             []Task
	VisitedToday      bool
	UpdatedAt         time.Time
}

type Service struct {
	db            *gorm.DB
	now           func() time.Time
	subscribersMu sync.Mutex
	subscribers   map[uuid.UUID]map[chan struct{}]struct{}
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:          db,
		now:         time.Now,
		subscribers: make(map[uuid.UUID]map[chan struct{}]struct{}),
	}
}

func (s *Service) Subscribe(userID uuid.UUID) (<-chan struct{}, func()) {
	updates := make(chan struct{}, 1)

	s.subscribersMu.Lock()

	if s.subscribers[userID] == nil {
		s.subscribers[userID] = make(map[chan struct{}]struct{})
	}

	s.subscribers[userID][updates] = struct{}{}
	s.subscribersMu.Unlock()

	var once sync.Once

	unsubscribe := func() {
		once.Do(func() {
			s.subscribersMu.Lock()

			delete(s.subscribers[userID], updates)

			if len(s.subscribers[userID]) == 0 {
				delete(s.subscribers, userID)
			}

			close(updates)

			s.subscribersMu.Unlock()
		})
	}

	return updates, unsubscribe
}

func (s *Service) Notify(userID uuid.UUID) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()

	for updates := range s.subscribers[userID] {
		select {
		case updates <- struct{}{}:
		default:
		}
	}
}

func (s *Service) Get(ctx context.Context, userID uuid.UUID) (DailyReport, error) {
	if userID == uuid.Nil {
		return DailyReport{}, ErrUserNotFound
	}

	now := s.now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)
	report := DailyReport{
		Date:      dayStart.Format(time.DateOnly),
		Rewards:   make([]Reward, 0),
		Tasks:     make([]Task, 0),
		UpdatedAt: dayStart,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User

		err := tx.Select("id").Where("id = ?", userID).First(&user).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}

		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		var rewards []models.Reward

		err = tx.
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, dayStart, dayEnd).
			Order("created_at ASC, id ASC").
			Find(&rewards).Error

		if err != nil {
			return fmt.Errorf("get rewards: %w", err)
		}

		for _, reward := range rewards {
			receivedAt := reward.CreatedAt.UTC()
			report.Rewards = append(report.Rewards, Reward{
				ID:         reward.ID,
				Type:       reward.Category,
				Title:      reward.Title,
				ExpiresAt:  reward.ExpiresAt.UTC(),
				ReceivedAt: receivedAt,
			})
			report.UpdatedAt = latestTime(report.UpdatedAt, receivedAt)
		}

		var taskRows []completedTaskRow

		err = tx.
			Table("user_daily_tasks AS assignments").
			Select(`assignments.id, definitions.type, definitions.title AS description,
				definitions.reward AS reward_leaves, assignments.claimed_at, assignments.completed_at`).
			Joins("JOIN daily_task_definitions AS definitions ON definitions.id = assignments.task_definition_id").
			Where("assignments.user_id = ? AND assignments.completed_at >= ? AND assignments.completed_at < ?", userID, dayStart, dayEnd).
			Order("assignments.completed_at ASC, definitions.slot ASC").
			Scan(&taskRows).Error

		if err != nil {
			return fmt.Errorf("get tasks: %w", err)
		}

		for _, row := range taskRows {
			completedAt := row.CompletedAt.UTC()

			report.Tasks = append(report.Tasks, Task{
				ID:            row.ID,
				Type:          row.Type,
				Description:   row.Description,
				RewardLeaves:  row.RewardLeaves,
				RewardClaimed: row.ClaimedAt != nil,
				CompletedAt:   completedAt,
			})

			report.UpdatedAt = latestTime(report.UpdatedAt, completedAt)

			if row.ClaimedAt != nil {
				report.UpdatedAt = latestTime(report.UpdatedAt, row.ClaimedAt.UTC())
			}
		}

		var login models.UserLogin

		err = tx.
			Where("user_id = ? AND activity_date >= ? AND activity_date < ?", userID, dayStart, dayEnd).
			First(&login).Error

		if err == nil {
			report.VisitedToday = true
			report.UpdatedAt = latestTime(report.UpdatedAt, login.CreatedAt.UTC())
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("get activity: %w", err)
		}

		var leafUpdatedAt time.Time

		report.LeavesEarnedToday, report.LevelUp, leafUpdatedAt, err = leafActivityForDay(tx, userID, dayStart, dayEnd)

		if err != nil {
			return err
		}

		report.UpdatedAt = latestTime(report.UpdatedAt, leafUpdatedAt)

		return nil
	})

	if err != nil {
		return DailyReport{}, fmt.Errorf("get daily report: %w", err)
	}

	return report, nil
}

func leafActivityForDay(tx *gorm.DB, userID uuid.UUID, dayStart, dayEnd time.Time) (int, *LevelUp, time.Time, error) {
	var transactions []models.LeafTransaction

	err := tx.
		Select("amount", "reason", "occurred_at").
		Where(`user_id = ? AND occurred_at >= ? AND occurred_at < ?
			AND reason IN (?, ?, ?)`, userID, dayStart, dayEnd,
			models.LeafReasonTaskReward, models.LeafReasonWeeklyLogin, models.LeafReasonLevelUp).
		Order("occurred_at ASC, id ASC").
		Find(&transactions).Error

	if err != nil {
		return 0, nil, time.Time{}, fmt.Errorf("get leaf activity: %w", err)
	}

	var earnedLeaves int64
	var levelUps []models.LeafTransaction
	var updatedAt time.Time

	for _, transaction := range transactions {
		switch transaction.Reason {
		case models.LeafReasonTaskReward, models.LeafReasonWeeklyLogin:
			if transaction.Amount > 0 {
				if earnedLeaves > int64(^uint64(0)>>1)-transaction.Amount {
					return 0, nil, time.Time{}, fmt.Errorf("calculate earned leaves: total overflows int64")
				}
				earnedLeaves += transaction.Amount
				updatedAt = latestTime(updatedAt, transaction.OccurredAt.UTC())
			}
		case models.LeafReasonLevelUp:
			levelUps = append(levelUps, transaction)
			updatedAt = latestTime(updatedAt, transaction.OccurredAt.UTC())
		}
	}

	if earnedLeaves > int64(^uint(0)>>1) {
		return 0, nil, time.Time{}, fmt.Errorf("calculate earned leaves: total %d overflows int", earnedLeaves)
	}

	if len(levelUps) == 0 {
		return int(earnedLeaves), nil, updatedAt, nil
	}

	var pet models.Pet

	if err := tx.Select("level").Where("user_id = ?", userID).First(&pet).Error; err != nil {
		return 0, nil, time.Time{}, fmt.Errorf("get pet level: %w", err)
	}

	fromLevel := pet.Level - len(levelUps)

	if fromLevel < 1 {
		return 0, nil, time.Time{}, fmt.Errorf("calculate level up: current level %d is inconsistent with %d level-up transactions", pet.Level, len(levelUps))
	}

	return int(earnedLeaves), &LevelUp{
		FromLevel:  fromLevel,
		ToLevel:    pet.Level,
		OccurredAt: levelUps[0].OccurredAt.UTC(),
	}, updatedAt, nil
}

func latestTime(current, candidate time.Time) time.Time {
	if candidate.After(current) {
		return candidate
	}

	return current
}

type completedTaskRow struct {
	ID           uuid.UUID
	Type         models.TaskType
	Description  string
	RewardLeaves int
	ClaimedAt    *time.Time
	CompletedAt  time.Time
}
