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

type RewardDefinition struct {
	Title    string
	Category models.RewardCategory
}

var defaultRewardDefinitions = [...]RewardDefinition{
	{Title: "1000 бонусов Авито", Category: models.RewardCategoryAvitoBonus},
	{Title: "Бесплатная доставка для трёх заказов", Category: models.RewardCategoryFreeDelivery},
	{Title: "Бесплатное продвижение объявления на 7 дней", Category: models.RewardCategoryFreePromotion},
	{Title: "Скидка 50% на продвижение объявления", Category: models.RewardCategoryPromotionDiscount},
	{Title: "Скидка 50% на Авито Доставку", Category: models.RewardCategoryDeliveryDiscount},
}

type Service struct {
	db           *gorm.DB
	pets         *pet.Service
	rewards      *rewards.Service
	now          func() time.Time
	selectReward func() (RewardDefinition, error)
}

func NewService(db *gorm.DB, petService *pet.Service, rewardService *rewards.Service) *Service {
	return &Service{
		db:           db,
		pets:         petService,
		rewards:      rewardService,
		now:          time.Now,
		selectReward: randomReward,
	}
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
		if userPet.Leaves < models.ChestOpeningLeavesCost {
			return ErrInsufficientLeaves
		}

		userPet.Leaves -= models.ChestOpeningLeavesCost
		if err := tx.Model(&userPet).Update("leaves", userPet.Leaves).Error; err != nil {
			return fmt.Errorf("spend leaves: %w", err)
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

		progress = pet.ProgressForPet(userPet, false)

		return nil
	})
	if err != nil {
		return models.Reward{}, err
	}

	service.pets.PublishProgress(userID, progress)

	return issuedReward, nil
}

func randomReward() (RewardDefinition, error) {
	index, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(defaultRewardDefinitions))))
	if err != nil {
		return RewardDefinition{}, err
	}

	return defaultRewardDefinitions[index.Int64()], nil
}
