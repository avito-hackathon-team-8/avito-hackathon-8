package events

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/tasks"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const LoginType = "LOGIN"

var (
	ErrInvalidEvent     = errors.New("invalid activity event")
	ErrEventConflict    = errors.New("event id was already used with different data")
	ErrUserNotFound     = errors.New("verified user not found")
	ErrEventOutsideTime = errors.New("event occurredAt is outside the allowed UTC window")
)

type Event struct {
	ID         uuid.UUID
	Type       string
	Count      int
	OccurredAt time.Time
}

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
}

type Service struct {
	db          *gorm.DB
	tasks       *tasks.Service
	now         func() time.Time
	dailyReport DailyReportNotifier
}

func NewService(db *gorm.DB, dailyReport DailyReportNotifier, taskService *tasks.Service) *Service {
	return &Service{db: db, tasks: taskService, now: time.Now, dailyReport: dailyReport}
}

func (service *Service) Record(ctx context.Context, userID uuid.UUID, batch []Event) error {
	if userID == uuid.Nil || len(batch) == 0 {
		return ErrInvalidEvent
	}

	var user models.User

	if err := service.db.WithContext(ctx).Where("id = ? AND verified = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	now := service.now().UTC()
	existingByID := make(map[uuid.UUID]models.ExternalEvent, len(batch))
	ids := make([]uuid.UUID, 0, len(batch))

	for _, event := range batch {
		if err := service.validateShape(event); err != nil {
			return err
		}

		ids = append(ids, event.ID)
	}

	var existing []models.ExternalEvent

	if err := service.db.WithContext(ctx).Where("id IN ?", ids).Find(&existing).Error; err != nil {
		return err
	}

	for _, record := range existing {
		existingByID[record.ID] = record
	}

	hasTaskEvent := false
	hasNewEvent := false

	for _, event := range batch {
		if record, exists := existingByID[event.ID]; exists {
			if record.PayloadHash != eventFingerprint(userID, event) {
				return ErrEventConflict
			}

			continue
		}

		if err := service.validateEvent(event, now); err != nil {
			return err
		}

		hasNewEvent = true

		if event.Type != LoginType {
			hasTaskEvent = true
		}
	}

	if !hasNewEvent {
		return nil
	}

	if hasTaskEvent {
		level, err := service.userLevel(ctx, service.db, userID)

		if err != nil {
			return err
		}

		if _, err := service.tasks.List(ctx, userID, level); err != nil {
			return err
		}
	}

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userLevel := 1

		if hasTaskEvent {
			var err error

			userLevel, err = service.userLevel(ctx, tx, userID)

			if err != nil {
				return err
			}
		}

		for _, event := range batch {
			hash := eventFingerprint(userID, event)
			record := models.ExternalEvent{
				ID: event.ID, UserID: userID, Type: event.Type, Count: event.Count,
				OccurredAt: event.OccurredAt.UTC(), PayloadHash: hash,
			}

			result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&record)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				var existing models.ExternalEvent

				if err := tx.First(&existing, "id = ?", event.ID).Error; err != nil {
					return err
				}

				if existing.PayloadHash != hash {
					return ErrEventConflict
				}

				continue
			}

			if event.Type == LoginType {
				activity := models.UserLogin{UserID: userID, ActivityDate: utcDate(event.OccurredAt), CreatedAt: now}

				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "user_id"}, {Name: "activity_date"}}, DoNothing: true,
				}).Create(&activity).Error; err != nil {
					return err
				}

				continue
			}

			if err := service.tasks.RecordEventsTx(tx, userID, []tasks.Event{{Type: models.TaskType(event.Type), Count: event.Count}}, userLevel, event.OccurredAt); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	service.dailyReport.Notify(userID)

	return nil
}

func (service *Service) validateEvent(event Event, now time.Time) error {
	if event.OccurredAt.After(now) {
		return ErrEventOutsideTime
	}

	eventDay := utcDate(event.OccurredAt)
	today := utcDate(now)

	if eventDay.After(today) {
		return ErrEventOutsideTime
	}

	if event.Type == LoginType {
		weekStart := today.AddDate(0, 0, -(int(today.Weekday())+6)%7)

		if eventDay.Before(weekStart) {
			return ErrEventOutsideTime
		}

		return nil
	}

	if !eventDay.Equal(today) {
		return ErrEventOutsideTime
	}

	return nil
}

func (service *Service) validateShape(event Event) error {
	if event.ID == uuid.Nil || event.Count < 1 || event.OccurredAt.IsZero() || (event.Type != LoginType && !models.IsKnownTaskType(models.TaskType(event.Type))) {
		return ErrInvalidEvent
	}

	return nil
}

func eventFingerprint(userID uuid.UUID, event Event) string {
	payload := fmt.Sprintf("%s|%s|%s|%d|%s", userID, event.ID, event.Type, event.Count, event.OccurredAt.UTC().Format(time.RFC3339Nano))
	digest := sha256.Sum256([]byte(payload))

	return hex.EncodeToString(digest[:])
}

func (service *Service) userLevel(ctx context.Context, db *gorm.DB, userID uuid.UUID) (int, error) {
	query := db.WithContext(ctx).Where("user_id = ?", userID)

	var userPet models.Pet

	if err := query.First(&userPet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 1, nil
		}

		return 0, fmt.Errorf("load pet level: %w", err)
	}

	return userPet.Level, nil
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()

	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
