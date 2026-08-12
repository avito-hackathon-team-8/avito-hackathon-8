package rewards

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidReward  = errors.New("invalid reward")
	ErrRewardNotFound = errors.New("active reward not found")
)

const (
	StatusActive   = "ACTIVE"
	StatusRedeemed = "REDEEMED"
	StatusExpired  = "EXPIRED"
)

var CategoryOrder = []models.RewardCategory{
	models.RewardCategoryAvitoBonus,
	models.RewardCategoryFreeDelivery,
	models.RewardCategoryFreePromotion,
	models.RewardCategoryPromotionDiscount,
	models.RewardCategoryDeliveryDiscount,
	models.RewardCategoryBowl,
	models.RewardCategoryBed,
}

type Grant struct {
	Title          string
	Category       models.RewardCategory
	Source         models.RewardSource
	ExpiresAt      time.Time
	LevelRewardID  *uuid.UUID
	ChestOpeningID *uuid.UUID
	ItemType       *models.ShopItemType
}

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
}

type Service struct {
	db          *gorm.DB
	now         func() time.Time
	dailyReport DailyReportNotifier
}

func NewService(db *gorm.DB, dailyReport DailyReportNotifier) *Service {
	return &Service{db: db, now: time.Now, dailyReport: dailyReport}
}

func (service *Service) Grant(ctx context.Context, userID uuid.UUID, grant Grant) (models.Reward, error) {
	reward, err := service.GrantTx(ctx, service.db, userID, grant)
	if err != nil {
		return models.Reward{}, err
	}

	service.dailyReport.Notify(userID)

	return reward, nil
}

func (service *Service) GrantTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, grant Grant) (models.Reward, error) {
	now := service.now().UTC()
	grant.Title = strings.TrimSpace(grant.Title)

	if grant.Title == "" || !validCategory(grant.Category) || !validGrantOrigin(grant) || !grant.ExpiresAt.After(now) {
		return models.Reward{}, ErrInvalidReward
	}

	reward := models.Reward{
		UserID:         userID,
		LevelRewardID:  grant.LevelRewardID,
		ChestOpeningID: grant.ChestOpeningID,
		ItemType:       grant.ItemType,
		Title:          grant.Title,
		Category:       grant.Category,
		Source:         grant.Source,
		ExpiresAt:      grant.ExpiresAt.UTC(),
	}

	if err := tx.WithContext(ctx).Create(&reward).Error; err != nil {
		return models.Reward{}, fmt.Errorf("grant reward: %w", err)
	}

	return reward, nil
}

func (service *Service) List(ctx context.Context, userID uuid.UUID) ([]models.Reward, error) {
	var rewards []models.Reward

	err := service.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("category ASC, created_at DESC").
		Find(&rewards).Error

	if err != nil {
		return nil, fmt.Errorf("list rewards: %w", err)
	}

	return rewards, nil
}

func (service *Service) Get(ctx context.Context, userID, rewardID uuid.UUID) (models.Reward, error) {
	var reward models.Reward

	err := service.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", rewardID, userID).
		First(&reward).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Reward{}, ErrRewardNotFound
	}

	if err != nil {
		return models.Reward{}, fmt.Errorf("get reward: %w", err)
	}

	return reward, nil
}

func (service *Service) Status(reward models.Reward) string {
	if reward.RedeemedAt != nil {
		return StatusRedeemed
	}

	if !reward.ExpiresAt.After(service.now().UTC()) {
		return StatusExpired
	}

	return StatusActive
}

func (service *Service) Redeem(ctx context.Context, userID, rewardID uuid.UUID) (models.Reward, error) {
	var reward models.Reward

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", rewardID, userID).
			First(&reward).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRewardNotFound
		}

		if err != nil {
			return err
		}

		now := service.now().UTC()

		if reward.Source == models.RewardSourceShop || reward.RedeemedAt != nil || !reward.ExpiresAt.After(now) {
			return ErrRewardNotFound
		}

		reward.RedeemedAt = &now

		return tx.Model(&reward).Update("redeemed_at", now).Error
	})

	if errors.Is(err, ErrRewardNotFound) {
		return models.Reward{}, err
	}

	if err != nil {
		return models.Reward{}, fmt.Errorf("redeem reward: %w", err)
	}

	return reward, nil
}

func validCategory(category models.RewardCategory) bool {
	for _, candidate := range CategoryOrder {
		if category == candidate {
			return true
		}
	}

	return false
}

func validSource(source models.RewardSource) bool {
	switch source {
	case models.RewardSourceLevel, models.RewardSourceChest, models.RewardSourceLeaderboard, models.RewardSourceShop:
		return true
	default:
		return false
	}
}

func validGrantOrigin(grant Grant) bool {
	if !validSource(grant.Source) {
		return false
	}

	switch grant.Source {
	case models.RewardSourceLevel:
		return grant.LevelRewardID != nil && grant.ChestOpeningID == nil && grant.ItemType == nil
	case models.RewardSourceChest:
		return grant.LevelRewardID == nil && grant.ChestOpeningID != nil && grant.ItemType == nil
	case models.RewardSourceLeaderboard:
		return grant.LevelRewardID == nil && grant.ChestOpeningID == nil && grant.ItemType == nil
	case models.RewardSourceShop:
		return grant.LevelRewardID == nil && grant.ChestOpeningID == nil && grant.ItemType != nil
	default:
		return false
	}
}
