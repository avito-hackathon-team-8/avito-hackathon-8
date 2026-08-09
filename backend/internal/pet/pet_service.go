package pet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MinPetLevel      = 1
	MaxPetLevel      = 10
	InitialPetLevel  = MaxPetLevel
	InitialPetLeaves = int64(1000)
)

var (
	ErrPetNotFound = errors.New("pet not found")
	ErrInvalidName = errors.New("pet name must contain from 1 to 35 characters")
)

type Update struct {
	UserID   uuid.UUID
	Progress Progress
}

type Service struct {
	db            *gorm.DB
	subscribersMu sync.Mutex
	subscribers   map[uuid.UUID]map[chan Update]struct{}
	levelClaims   *LevelClaimsService
	dailyReport   DailyReportNotifier
	now           func() time.Time
}

func NewService(db *gorm.DB, dailyReport ...DailyReportNotifier) *Service {
	service := &Service{
		db:          db,
		subscribers: make(map[uuid.UUID]map[chan Update]struct{}),
		now:         time.Now,
	}

	if len(dailyReport) > 0 {
		service.dailyReport = dailyReport[0]
	}

	return service
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
			Level:  InitialPetLevel,
			Leaves: InitialPetLeaves,
		}

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
	return service.Credit(ctx, Credit{
		UserID:       userID,
		Amount:       amount,
		Reason:       models.LeafReasonTaskReward,
		OperationKey: fmt.Sprintf("pet:add:%s", uuid.NewString()),
	})
}

func (service *Service) AddLeavesTx(tx *gorm.DB, userID uuid.UUID, amount int64) (Progress, error) {
	return service.CreditTx(tx, Credit{
		UserID:       userID,
		Amount:       amount,
		Reason:       models.LeafReasonTaskReward,
		OperationKey: fmt.Sprintf("pet:add:%s", uuid.NewString()),
	})
}

func (service *Service) PublishProgress(userID uuid.UUID, progress Progress) {
	service.publish(Update{UserID: userID, Progress: progress})
}
