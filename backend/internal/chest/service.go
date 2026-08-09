package chest

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/reward_catalog"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ChestRewardLifetime = 30 * 24 * time.Hour

var (
	ErrPetNotFound        = errors.New("pet not found")
	ErrChestLevelRequired = errors.New("chests are available from pet level 10")
	ErrInsufficientLeaves = errors.New("insufficient leaves to open chest")
)

type RewardDefinition = reward_catalog.ChestRewardDefinition

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
}

type Service struct {
	db           *gorm.DB
	pets         *pet.Service
	rewards      *rewards.Service
	now          func() time.Time
	selectReward func() (RewardDefinition, error)
	dailyReport  DailyReportNotifier
}

func NewService(db *gorm.DB, dailyReport DailyReportNotifier, petService *pet.Service, rewardService *rewards.Service, definitions []RewardDefinition) *Service {
	service := &Service{
		db:          db,
		dailyReport: dailyReport,
		pets:        petService,
		rewards:     rewardService,
		now:         time.Now,
	}
	service.selectReward = randomReward(definitions)

	return service
}

func (service *Service) Open(ctx context.Context, userID uuid.UUID) (models.Reward, error) {
	if userID == uuid.Nil {
		return models.Reward{}, ErrPetNotFound
	}

	var issuedReward models.Reward
	var progress pet.Progress

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userPet models.Pet

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&userPet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPetNotFound
		}

		if err != nil {
			return fmt.Errorf("lock pet: %w", err)
		}

		if userPet.Level < pet.MaxPetLevel {
			return ErrChestLevelRequired
		}

		now := service.now().UTC()
		opening := models.ChestOpening{
			UserID:      userID,
			LeavesSpent: models.ChestOpeningLeavesCost,
			OpenedAt:    now,
		}

		if err := tx.Create(&opening).Error; err != nil {
			return fmt.Errorf("create chest opening: %w", err)
		}
		progress, err = service.pets.DebitTx(tx, pet.Debit{
			UserID:       userID,
			Amount:       models.ChestOpeningLeavesCost,
			Reason:       models.LeafReasonChestPurchase,
			OperationKey: fmt.Sprintf("chest:%s", opening.ID),
			OccurredAt:   now,
		})
		if errors.Is(err, pet.ErrInsufficientLeaves) {
			return ErrInsufficientLeaves
		}
		if err != nil {
			return fmt.Errorf("spend leaves: %w", err)
		}

		definition, err := service.selectReward()

		if err != nil {
			return fmt.Errorf("select chest reward: %w", err)
		}

		issuedReward, err = service.rewards.GrantTx(ctx, tx, userID, rewards.Grant{
			Title:          definition.Title,
			Category:       definition.Category,
			Source:         models.RewardSourceChest,
			ExpiresAt:      now.Add(ChestRewardLifetime),
			ChestOpeningID: &opening.ID,
		})

		if err != nil {
			return fmt.Errorf("issue chest reward: %w", err)
		}

		return nil
	})

	if err != nil {
		return models.Reward{}, err
	}

	service.pets.PublishProgress(userID, progress)
	service.dailyReport.Notify(userID)

	return issuedReward, nil
}

func randomReward(definitions []RewardDefinition) func() (RewardDefinition, error) {
	return func() (RewardDefinition, error) {
		if len(definitions) == 0 {
			return RewardDefinition{}, errors.New("chest rewards are not configured")
		}

		index, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(definitions))))
		if err != nil {
			return RewardDefinition{}, err
		}

		return definitions[index.Int64()], nil
	}
}
