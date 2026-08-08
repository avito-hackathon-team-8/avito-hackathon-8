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
		UpdatedAt: now,
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
		}

		var taskRows []completedTaskRow

		err = tx.
			Table("user_task_progresses AS progress").
			Select(`tasks.id, tasks.type, tasks.description, tasks.reward_leaves,
				progress.claimed_at, progress.completed_at`).
			Joins("JOIN tasks ON tasks.id = progress.task_id").
			Where("progress.user_id = ? AND progress.completed_at >= ? AND progress.completed_at < ?", userID, dayStart, dayEnd).
			Order("progress.completed_at ASC, tasks.slot ASC").
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

			if row.ClaimedAt != nil && !row.ClaimedAt.Before(dayStart) && row.ClaimedAt.Before(dayEnd) {
				report.LeavesEarnedToday += row.RewardLeaves
			}
		}

		var login models.UserLogin

		err = tx.
			Where("user_id = ? AND activity_date >= ? AND activity_date < ?", userID, dayStart, dayEnd).
			First(&login).Error

		if err == nil {
			report.VisitedToday = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("get activity: %w", err)
		}

		if report.VisitedToday {
			var loginClaim models.WeeklyLoginClaim

			err = tx.
				Where("user_id = ? AND claim_date >= ? AND claim_date < ?", userID, dayStart, dayEnd).
				First(&loginClaim).Error

			if err == nil {
				report.LeavesEarnedToday += loginClaim.RewardLeaves
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("get login reward: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return DailyReport{}, fmt.Errorf("get daily report: %w", err)
	}

	return report, nil
}

type completedTaskRow struct {
	ID           uuid.UUID
	Type         models.TaskType
	Description  string
	RewardLeaves int
	ClaimedAt    *time.Time
	CompletedAt  time.Time
}
