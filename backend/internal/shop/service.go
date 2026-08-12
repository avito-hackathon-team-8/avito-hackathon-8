package shop

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrItemNotFound            = errors.New("shop item not found")
	ErrPetNotFound             = errors.New("pet not found")
	ErrLevelRequired           = errors.New("pet level is too low for shop item")
	ErrInsufficientLeaves      = errors.New("insufficient leaves to purchase shop item")
	ErrReplacementConfirmation = errors.New("active shop item replacement requires confirmation")
)

type Purchase struct {
	ItemID             string
	ConfirmReplacement bool
}

type Service struct {
	db      *gorm.DB
	pets    *pet.Service
	rewards *rewards.Service
	catalog Catalog
	now     func() time.Time
}

func NewService(db *gorm.DB, petService *pet.Service, rewardService *rewards.Service, catalog Catalog) *Service {
	return &Service{db: db, pets: petService, rewards: rewardService, catalog: catalog, now: time.Now}
}

func (service *Service) Items() []models.ShopItem {
	return service.catalog.Items()
}

func (service *Service) Purchase(ctx context.Context, userID uuid.UUID, purchase Purchase) (models.Reward, pet.Progress, error) {
	if userID == uuid.Nil {
		return models.Reward{}, pet.Progress{}, ErrPetNotFound
	}

	item, exists := service.catalog.Item(purchase.ItemID)
	if !exists {
		return models.Reward{}, pet.Progress{}, ErrItemNotFound
	}

	var reward models.Reward
	var progress pet.Progress
	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userPet models.Pet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&userPet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPetNotFound
			}
			return fmt.Errorf("lock pet: %w", err)
		}
		if userPet.Level < item.RequiredLevel {
			return ErrLevelRequired
		}

		now := service.now().UTC()
		var active models.Reward
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND source = ? AND category = ? AND expires_at > ?", userID, models.RewardSourceShop, item.Category, now).
			Order("expires_at DESC").
			First(&active).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load active shop reward: %w", err)
		}
		hasActive := err == nil

		if hasActive && active.ItemType != nil && *active.ItemType != item.Type && !purchase.ConfirmReplacement {
			return ErrReplacementConfirmation
		}

		progress, err = service.pets.DebitTx(tx, pet.Debit{
			UserID:       userID,
			Amount:       item.PriceLeaves,
			Reason:       models.LeafReasonShopPurchase,
			OperationKey: fmt.Sprintf("shop:%s", uuid.NewString()),
			OccurredAt:   now,
		})
		if errors.Is(err, pet.ErrInsufficientLeaves) {
			return ErrInsufficientLeaves
		}
		if err != nil {
			return fmt.Errorf("spend leaves: %w", err)
		}

		if hasActive && active.ItemType != nil && *active.ItemType == item.Type {
			expiresAt := active.ExpiresAt.AddDate(0, 0, item.DurationDays)
			if err := tx.Model(&active).Update("expires_at", expiresAt).Error; err != nil {
				return fmt.Errorf("extend shop reward: %w", err)
			}
			active.ExpiresAt = expiresAt
			reward = active
			return nil
		}

		if hasActive {
			if err := tx.Model(&active).Update("expires_at", now).Error; err != nil {
				return fmt.Errorf("expire replaced shop reward: %w", err)
			}
		}

		itemType := item.Type
		reward, err = service.rewards.GrantTx(ctx, tx, userID, rewards.Grant{
			Title:     item.Title,
			Category:  item.Category,
			Source:    models.RewardSourceShop,
			ExpiresAt: now.AddDate(0, 0, item.DurationDays),
			ItemType:  &itemType,
		})
		if err != nil {
			return fmt.Errorf("issue shop reward: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.Reward{}, pet.Progress{}, err
	}

	service.pets.PublishProgress(userID, progress)
	return reward, progress, nil
}
