package pet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/leaves"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MinPetLevel = leaves.MinPetLevel
	MaxPetLevel = leaves.MaxPetLevel
	maxInt64    = int64(1<<63 - 1)
)

var (
	ErrPetNotFound    = errors.New("pet not found")
	ErrInvalidName    = errors.New("pet name must contain from 1 to 35 characters")
	ErrInvalidLeaves  = leaves.ErrInvalidAmount
	ErrLeavesOverflow = leaves.ErrLeavesOverflow
)

type Progress = leaves.Progress

type Update struct {
	UserID   uuid.UUID
	Progress Progress
}

type Service struct {
	db            *gorm.DB
	subscribersMu sync.Mutex
	subscribers   map[uuid.UUID]map[chan Update]struct{}
	levelClaims   *LevelClaimsService
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:          db,
		subscribers: make(map[uuid.UUID]map[chan Update]struct{}),
	}
}

func (service *Service) SetLevelClaimsService(levelClaims *LevelClaimsService) {
	service.levelClaims = levelClaims
}

func (service *Service) Subscribe(userID uuid.UUID) (<-chan Update, func()) {
	updates := make(chan Update, 8)
	service.subscribersMu.Lock()
	if service.subscribers[userID] == nil {
		service.subscribers[userID] = make(map[chan Update]struct{})
	}
	service.subscribers[userID][updates] = struct{}{}
	service.subscribersMu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			service.subscribersMu.Lock()
			delete(service.subscribers[userID], updates)
			if len(service.subscribers[userID]) == 0 {
				delete(service.subscribers, userID)
			}
			close(updates)
			service.subscribersMu.Unlock()
		})
	}

	return updates, unsubscribe
}

func (service *Service) publish(update Update) {
	service.subscribersMu.Lock()
	defer service.subscribersMu.Unlock()

	for updates := range service.subscribers[update.UserID] {
		select {
		case updates <- update:
		default:
			select {
			case <-updates:
			default:
			}
			select {
			case updates <- update:
			default:
			}
		}
	}
}

func (service *Service) GetOrCreate(ctx context.Context, userID uuid.UUID) (models.Pet, error) {
	if userID == uuid.Nil {
		return models.Pet{}, ErrPetNotFound
	}

	var pet models.Pet

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newPet := models.Pet{
			UserID: userID,
		}

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&newPet).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", userID).First(&pet).Error; err != nil {
			return err
		}

		if service.levelClaims != nil {
			if err := service.levelClaims.openReachedRewards(tx, userID, pet.Level); err != nil {
				return err
			}
		}

		return nil
	})

	if errors.Is(err, ErrPetNotFound) {
		return models.Pet{}, err
	}

	if err != nil {
		return models.Pet{}, fmt.Errorf("get or create pet: %w", err)
	}

	return pet, nil
}

func (service *Service) UpdateName(ctx context.Context, userID uuid.UUID, name string) (models.Pet, error) {
	name = strings.TrimSpace(name)

	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 35 {
		return models.Pet{}, ErrInvalidName
	}

	var pet models.Pet

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&pet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPetNotFound
		}

		if err != nil {
			return err
		}

		if err := tx.Model(&pet).Update("name", name).Error; err != nil {
			return err
		}

		pet.Name = name

		return nil
	})

	if errors.Is(err, ErrPetNotFound) {
		return models.Pet{}, err
	}

	if err != nil {
		return models.Pet{}, fmt.Errorf("update pet name: %w", err)
	}

	return pet, nil
}

func (service *Service) AddLeaves(ctx context.Context, userID uuid.UUID, amount int64) (Progress, error) {
	if amount <= 0 {
		return Progress{}, ErrInvalidLeaves
	}

	var progress Progress
	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transactionProgress, err := service.addLeavesTx(tx, userID, amount)
		progress = transactionProgress

		return err
	})

	if errors.Is(err, ErrInvalidLeaves) ||
		errors.Is(err, ErrPetNotFound) ||
		errors.Is(err, ErrLeavesOverflow) {
		return Progress{}, err
	}

	if err != nil {
		return Progress{}, fmt.Errorf("add leaves: %w", err)
	}

	service.publish(Update{UserID: userID, Progress: progress})

	return progress, nil
}

func (service *Service) AddLeavesTx(tx *gorm.DB, userID uuid.UUID, amount int64) (Progress, error) {
	if amount <= 0 {
		return Progress{}, ErrInvalidLeaves
	}

	return service.addLeavesTx(tx, userID, amount)
}

func (service *Service) addLeavesTx(tx *gorm.DB, userID uuid.UUID, amount int64) (Progress, error) {
	var pet models.Pet

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&pet).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Progress{}, ErrPetNotFound
	}

	if err != nil {
		return Progress{}, err
	}

	if pet.Leaves < 0 || pet.Leaves > maxInt64-amount {
		return Progress{}, ErrLeavesOverflow
	}

	newLevel, remainingLeaves := leaves.ApplyLevelUps(pet.Level, pet.Leaves+amount)
	oldLevel := pet.Level

	pet.Leaves = remainingLeaves
	pet.Level = newLevel

	if err := tx.Model(&pet).Updates(map[string]any{
		"leaves": pet.Leaves,
		"level":  newLevel,
	}).Error; err != nil {
		return Progress{}, err
	}
	if newLevel > oldLevel && service.levelClaims != nil {
		if err := service.levelClaims.openReachedRewards(tx, userID, newLevel); err != nil {
			return Progress{}, err
		}
	}

	return ProgressForPet(pet, newLevel > oldLevel), nil
}

func (service *Service) PublishProgress(userID uuid.UUID, progress Progress) {
	service.publish(Update{UserID: userID, Progress: progress})
}

func ProgressForPet(pet models.Pet, levelUp bool) Progress {
	return leaves.ProgressForPet(pet, levelUp)
}
