package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	InitialHappiness = 100.0
	DecayDuration    = 72 * time.Hour
	CareCooldown     = 6 * time.Hour
	KafkaTopic       = "pet-state-events-v1"
)

type CareType string

const (
	CareStroke CareType = "STROKE"
	CareFeed   CareType = "FEED"
)

var (
	ErrInvalidCareType       = errors.New("invalid care type")
	ErrCareCooldown          = errors.New("pet care action is on cooldown")
	ErrHappinessFull         = errors.New("pet happiness is already full")
	ErrIdempotencyConflict   = errors.New("idempotency key was already used for another operation")
	ErrInvalidIdempotencyKey = errors.New("idempotency key is too long")
)

type CooldownError struct{ NextAvailableAt time.Time }

func (err *CooldownError) Error() string { return ErrCareCooldown.Error() }
func (err *CooldownError) Unwrap() error { return ErrCareCooldown }

type PetState struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Happiness float64   `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type PetStateAction struct {
	UserID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ActionType      CareType  `gorm:"primaryKey;column:action_type"`
	NextAvailableAt time.Time `gorm:"not null"`
}

type OutboxEvent struct {
	EventID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Topic       string    `gorm:"not null"`
	EventKey    string    `gorm:"not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	PublishedAt *time.Time
	LeaseToken  *uuid.UUID `gorm:"type:uuid"`
	LeaseUntil  *time.Time
}

type CareIdempotency struct {
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	IdempotencyKey string    `gorm:"size:128;primaryKey"`
	ActionType     CareType  `gorm:"not null"`
	Response       []byte    `gorm:"type:jsonb;not null"`
	CreatedAt      time.Time `gorm:"not null"`
}

func (PetState) TableName() string        { return "pet_states" }
func (PetStateAction) TableName() string  { return "pet_state_actions" }
func (OutboxEvent) TableName() string     { return "pet_state_outbox" }
func (CareIdempotency) TableName() string { return "pet_state_care_idempotency" }

type Snapshot struct {
	Happiness             float64    `json:"happiness"`
	CalculatedAt          time.Time  `json:"calculatedAt"`
	DecaysToZeroAt        time.Time  `json:"decaysToZeroAt"`
	StrokeNextAvailableAt *time.Time `json:"strokeNextAvailableAt"`
	FeedNextAvailableAt   *time.Time `json:"feedNextAvailableAt"`
	HappinessMultiplier   float64    `json:"happinessMultiplier"`
}

type ChangedEvent struct {
	EventID    uuid.UUID `json:"eventId"`
	Version    int       `json:"version"`
	Type       string    `json:"type"`
	UserID     uuid.UUID `json:"userId"`
	OccurredAt time.Time `json:"occurredAt"`
	Data       Snapshot  `json:"data"`
}

type Service struct {
	db    *gorm.DB
	now   func() time.Time
	topic string
}

func NewService(db *gorm.DB, topic string) *Service {
	if topic == "" {
		topic = KafkaTopic
	}

	return &Service{db: db, now: time.Now, topic: topic}
}

func (service *Service) Get(ctx context.Context, userID uuid.UUID) (Snapshot, error) {
	now := service.now().UTC()
	petState, err := loadState(service.db.WithContext(ctx), userID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			petState, err = ensureState(tx, userID, now)
			return err
		})
	}

	if err != nil {
		return Snapshot{}, fmt.Errorf("load pet state: %w", err)
	}

	return snapshot(service.db.WithContext(ctx), petState, now)
}

func (service *Service) Care(ctx context.Context, userID uuid.UUID, careType CareType, idempotencyKey string) (Snapshot, error) {
	delta, ok := careDelta(careType)

	if !ok {
		return Snapshot{}, ErrInvalidCareType
	}

	if len(idempotencyKey) > 128 {
		return Snapshot{}, ErrInvalidIdempotencyKey
	}

	now := service.now().UTC()

	var result Snapshot

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := ensureState(tx, userID, now); err != nil {
			return err
		}

		var petState PetState

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&petState).Error; err != nil {
			return fmt.Errorf("lock pet state: %w", err)
		}

		if idempotencyKey != "" {
			stored, found, err := loadIdempotentResult(tx, userID, idempotencyKey, careType)
			if err != nil {
				return err
			}
			if found {
				result = stored
				return nil
			}
		}

		current := DecayedHappiness(petState.Happiness, petState.UpdatedAt, now)

		if current >= InitialHappiness-1e-9 {
			return ErrHappinessFull
		}

		var action PetStateAction

		err := tx.Where("user_id = ? AND action_type = ?", userID, careType).First(&action).Error

		if err == nil && now.Before(action.NextAvailableAt) {
			return &CooldownError{NextAvailableAt: action.NextAvailableAt.UTC()}
		}

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load care cooldown: %w", err)
		}

		petState.Happiness = math.Min(InitialHappiness, current+delta)
		petState.UpdatedAt = now

		if err := tx.Model(&PetState{}).Where("user_id = ?", userID).Updates(map[string]any{
			"happiness":  petState.Happiness,
			"updated_at": now,
		}).Error; err != nil {
			return fmt.Errorf("update pet state: %w", err)
		}

		action = PetStateAction{UserID: userID, ActionType: careType, NextAvailableAt: now.Add(CareCooldown)}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "action_type"}},
			DoUpdates: clause.AssignmentColumns([]string{"next_available_at"}),
		}).Create(&action).Error; err != nil {
			return fmt.Errorf("save care cooldown: %w", err)
		}

		var snapshotErr error

		result, snapshotErr = snapshot(tx, petState, now)

		if snapshotErr != nil {
			return snapshotErr
		}

		event := ChangedEvent{EventID: uuid.New(), Version: 1, Type: "PET_STATE_CHANGED", UserID: userID, OccurredAt: now, Data: result}
		payload, err := json.Marshal(event)

		if err != nil {
			return fmt.Errorf("encode state event: %w", err)
		}

		if err := tx.Create(&OutboxEvent{EventID: event.EventID, Topic: service.topic, EventKey: userID.String(), Payload: payload, CreatedAt: now}).Error; err != nil {
			return fmt.Errorf("write state outbox: %w", err)
		}

		if idempotencyKey != "" {
			encodedResult, err := json.Marshal(result)

			if err != nil {
				return fmt.Errorf("encode idempotent result: %w", err)
			}

			entry := CareIdempotency{UserID: userID, IdempotencyKey: idempotencyKey, ActionType: careType, Response: encodedResult, CreatedAt: now}

			if err := tx.Create(&entry).Error; err != nil {
				return fmt.Errorf("save idempotent result: %w", err)
			}
		}

		return nil
	})

	return result, err
}

func loadState(db *gorm.DB, userID uuid.UUID) (PetState, error) {
	var petState PetState

	err := db.Where("user_id = ?", userID).First(&petState).Error

	return petState, err
}

func loadIdempotentResult(tx *gorm.DB, userID uuid.UUID, key string, careType CareType) (Snapshot, bool, error) {
	var entry CareIdempotency

	err := tx.Where("user_id = ? AND idempotency_key = ?", userID, key).First(&entry).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Snapshot{}, false, nil
	}

	if err != nil {
		return Snapshot{}, false, fmt.Errorf("load idempotent result: %w", err)
	}

	if entry.ActionType != careType {
		return Snapshot{}, false, ErrIdempotencyConflict
	}

	var result Snapshot

	if err := json.Unmarshal(entry.Response, &result); err != nil {
		return Snapshot{}, false, fmt.Errorf("decode idempotent result: %w", err)
	}

	return result, true, nil
}

func DecayedHappiness(value float64, updatedAt, now time.Time) float64 {
	if !now.After(updatedAt) {
		return clamp(value)
	}

	return clamp(value - now.Sub(updatedAt).Seconds()*(InitialHappiness/DecayDuration.Seconds()))
}

func ensureState(tx *gorm.DB, userID uuid.UUID, now time.Time) (PetState, error) {
	created := PetState{UserID: userID, Happiness: InitialHappiness, UpdatedAt: now}

	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&created).Error; err != nil {
		return PetState{}, fmt.Errorf("ensure pet state: %w", err)
	}

	var current PetState

	if err := tx.Where("user_id = ?", userID).First(&current).Error; err != nil {
		return PetState{}, fmt.Errorf("load pet state: %w", err)
	}

	return current, nil
}

func snapshot(tx *gorm.DB, petState PetState, now time.Time) (Snapshot, error) {
	happiness := DecayedHappiness(petState.Happiness, petState.UpdatedAt, now)

	result := Snapshot{
		Happiness:           math.Round(happiness*100) / 100,
		CalculatedAt:        now,
		DecaysToZeroAt:      now.Add(time.Duration(happiness / InitialHappiness * float64(DecayDuration))),
		HappinessMultiplier: math.Round((0.5+happiness/100)*1000) / 1000,
	}

	var actions []PetStateAction

	if err := tx.Where("user_id = ?", petState.UserID).Find(&actions).Error; err != nil {
		return Snapshot{}, fmt.Errorf("load care cooldowns: %w", err)
	}

	for _, action := range actions {
		if !action.NextAvailableAt.After(now) {
			continue
		}

		value := action.NextAvailableAt.UTC()

		switch action.ActionType {
		case CareStroke:
			result.StrokeNextAvailableAt = &value
		case CareFeed:
			result.FeedNextAvailableAt = &value
		}
	}

	return result, nil
}

func careDelta(careType CareType) (float64, bool) {
	switch careType {
	case CareStroke:
		return 20, true
	case CareFeed:
		return 35, true
	default:
		return 0, false
	}
}

func clamp(value float64) float64 { return math.Max(0, math.Min(InitialHappiness, value)) }
