package pet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MinPetLevel = 1
	MaxPetLevel = 10
	maxInt64    = int64(1<<63 - 1)
)

var (
	ErrPetNotFound    = errors.New("pet not found")
	ErrInvalidName    = errors.New("pet name must contain from 1 to 35 characters")
	ErrInvalidLeaves  = errors.New("leaves amount must be positive")
	ErrLeavesOverflow = errors.New("leaves amount overflows int64")
)

var levelTargets = [...]int64{0, 0, 100, 230, 390, 580, 810, 1090, 1430, 1850, 2400}

type Progress struct {
	Name                  string
	Level                 int
	Leaves                int64
	NextLevelTargetLeaves int64
	LevelUp               bool
}

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
	return &Service{db: db, subscribers: make(map[uuid.UUID]map[chan Update]struct{})}
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
		newPet := models.Pet{UserID: userID}

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&newPet).Error; err != nil {
			return fmt.Errorf("create pet: %w", err)
		}

		if err := tx.Where("user_id = ?", userID).First(&pet).Error; err != nil {
			return fmt.Errorf("load pet: %w", err)
		}

		if service.levelClaims != nil {
			if err := service.levelClaims.openReachedRewards(tx, userID, pet.Level); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return models.Pet{}, err
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
			return fmt.Errorf("lock pet: %w", err)
		}

		if err := tx.Model(&pet).Update("name", name).Error; err != nil {
			return fmt.Errorf("update pet name: %w", err)
		}

		pet.Name = name

		return nil
	})

	if err != nil {
		return models.Pet{}, err
	}

	return pet, nil
}

func (service *Service) AddLeaves(ctx context.Context, userID uuid.UUID, amount int64) (Progress, error) {
	if amount <= 0 {
		return Progress{}, ErrInvalidLeaves
	}

	var progress Progress

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		progress, err = service.addLeavesTx(tx, userID, amount)
		return err
	})

	if err != nil {
		return Progress{}, err
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
		return Progress{}, fmt.Errorf("lock pet: %w", err)
	}

	if pet.Leaves < 0 || pet.Leaves > maxInt64-amount {
		return Progress{}, ErrLeavesOverflow
	}

	newLevel, remainingLeaves := applyLevelUps(pet.Level, pet.Leaves+amount)
	oldLevel := pet.Level

	pet.Leaves = remainingLeaves
	pet.Level = newLevel

	err = tx.Model(&pet).Updates(map[string]any{
		"leaves": pet.Leaves,
		"level":  newLevel,
	}).Error

	if err != nil {
		return Progress{}, fmt.Errorf("save pet progress: %w", err)
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

func levelCost(level int) int64 {
	if level < MinPetLevel || level >= MaxPetLevel {
		return 0
	}

	return levelTargets[level+1] - levelTargets[level]
}

func applyLevelUps(level int, leaves int64) (int, int64) {
	for level < MaxPetLevel {
		cost := levelCost(level)
		if leaves < cost {
			break
		}
		leaves -= cost
		level++
	}

	return level, leaves
}

func ProgressForPet(pet models.Pet, levelUp bool) Progress {
	if pet.Level >= MaxPetLevel {
		return Progress{
			Name:                  pet.Name,
			Level:                 MaxPetLevel,
			Leaves:                pet.Leaves,
			NextLevelTargetLeaves: 0,
			LevelUp:               levelUp,
		}
	}

	return Progress{
		Name:                  pet.Name,
		Level:                 pet.Level,
		Leaves:                pet.Leaves,
		NextLevelTargetLeaves: levelCost(pet.Level),
		LevelUp:               levelUp,
	}
}
