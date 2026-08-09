package pet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/reward_catalog"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	LevelRewardClaimWindow     = 72 * time.Hour
	DefaultLevelRewardLifetime = LevelRewardClaimWindow
)

var (
	ErrLevelRewardNotFound       = errors.New("level reward not found")
	ErrLevelRewardLocked         = errors.New("level reward is locked")
	ErrLevelRewardFrozen         = errors.New("level reward claim window has expired")
	ErrLevelRewardAlreadyClaimed = errors.New("level reward already claimed")
)

type LevelRewardDefinition = reward_catalog.LevelRewardDefinition

type LevelRewardItem struct {
	Level     int
	Status    models.LevelRewardStatus
	Reward    models.LevelReward
	ExpiresAt *time.Time
}

type LevelClaimResult struct {
	Level  int
	Status models.LevelRewardStatus
	Reward models.Reward
}

type LevelClaimsService struct {
	db             *gorm.DB
	rewards        *rewards.Service
	now            func() time.Time
	claimWindow    time.Duration
	rewardLifetime time.Duration
	dailyReport    DailyReportNotifier
	definitions    []LevelRewardDefinition
}

func NewLevelClaimsService(db *gorm.DB, dailyReport DailyReportNotifier, rewardService *rewards.Service, definitions []LevelRewardDefinition) *LevelClaimsService {
	if rewardService == nil {
		rewardService = rewards.NewService(db, dailyReport)
	}

	return &LevelClaimsService{
		db:             db,
		rewards:        rewardService,
		now:            time.Now,
		claimWindow:    LevelRewardClaimWindow,
		rewardLifetime: DefaultLevelRewardLifetime,
		dailyReport:    dailyReport,
		definitions:    append([]LevelRewardDefinition(nil), definitions...),
	}
}

func (service *LevelClaimsService) GetLevels(ctx context.Context, userID uuid.UUID) ([]LevelRewardItem, error) {
	if userID == uuid.Nil {
		return nil, ErrPetNotFound
	}

	var items []LevelRewardItem
	now := service.now().UTC()

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userPet, err := levelClaimsPet(tx, userID)
		if err != nil {
			return err
		}

		levelRewards, err := service.ensureLevelRewards(tx, userID, userPet.Level, now)
		if err != nil {
			return err
		}

		items = make([]LevelRewardItem, 0, len(levelRewards))
		for _, levelReward := range levelRewards {
			status := levelRewardStatus(levelReward, userPet.Level, now)
			items = append(items, LevelRewardItem{
				Level:     levelReward.Level,
				Status:    status,
				Reward:    levelReward,
				ExpiresAt: levelRewardExpiresAt(levelReward, status),
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (service *LevelClaimsService) openReachedRewards(tx *gorm.DB, userID uuid.UUID, userLevel int) error {
	_, err := service.ensureLevelRewards(tx, userID, userLevel, service.now().UTC())
	return err
}

func (service *LevelClaimsService) Claim(ctx context.Context, userID, levelRewardID uuid.UUID) (LevelClaimResult, error) {
	if userID == uuid.Nil {
		return LevelClaimResult{}, ErrPetNotFound
	}
	if levelRewardID == uuid.Nil {
		return LevelClaimResult{}, ErrLevelRewardNotFound
	}

	var result LevelClaimResult
	now := service.now().UTC()

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userPet, err := levelClaimsPet(tx, userID)
		if err != nil {
			return err
		}

		if _, err := service.ensureLevelRewards(tx, userID, userPet.Level, now); err != nil {
			return err
		}

		var levelReward models.LevelReward
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", levelRewardID, userID).
			First(&levelReward).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLevelRewardNotFound
		}
		if err != nil {
			return fmt.Errorf("lock level reward: %w", err)
		}

		switch levelRewardStatus(levelReward, userPet.Level, now) {
		case models.LevelRewardStatusLocked:
			return ErrLevelRewardLocked
		case models.LevelRewardStatusFrozen:
			return ErrLevelRewardFrozen
		case models.LevelRewardStatusClaimed:
			return ErrLevelRewardAlreadyClaimed
		}

		expiresAt := now.Add(service.rewardLifetime)
		issuedReward, err := service.rewards.GrantTx(ctx, tx, userID, rewards.Grant{
			Title:         levelReward.Title,
			Category:      levelReward.Category,
			Source:        models.RewardSourceLevel,
			ExpiresAt:     expiresAt,
			LevelRewardID: &levelReward.ID,
		})
		if err != nil {
			return fmt.Errorf("issue level reward: %w", err)
		}

		if err := tx.Model(&levelReward).Update("claimed_at", now).Error; err != nil {
			return fmt.Errorf("mark level reward claimed: %w", err)
		}

		result = LevelClaimResult{
			Level:  levelReward.Level,
			Status: models.LevelRewardStatusClaimed,
			Reward: issuedReward,
		}
		return nil
	})
	if err != nil {
		return LevelClaimResult{}, err
	}

	service.dailyReport.Notify(userID)

	return result, nil
}

func (service *LevelClaimsService) ensureLevelRewards(tx *gorm.DB, userID uuid.UUID, userLevel int, now time.Time) ([]models.LevelReward, error) {
	for _, definition := range service.definitions {
		levelReward := models.LevelReward{
			UserID:      userID,
			Level:       definition.Level,
			Title:       definition.Title,
			Description: definition.Description,
			Category:    definition.Category,
		}
		if definition.Level <= userLevel {
			expiresAt := now.Add(service.claimWindow)
			levelReward.ClaimExpiresAt = &expiresAt
		}

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&levelReward).Error; err != nil {
			return nil, fmt.Errorf("create level reward: %w", err)
		}
	}

	var levelRewards []models.LevelReward
	if err := tx.Where("user_id = ?", userID).Order("level ASC").Find(&levelRewards).Error; err != nil {
		return nil, fmt.Errorf("list level rewards: %w", err)
	}

	for i := range levelRewards {
		levelReward := &levelRewards[i]
		if levelReward.Level > userLevel || levelReward.ClaimExpiresAt != nil || levelReward.ClaimedAt != nil {
			continue
		}

		expiresAt := now.Add(service.claimWindow)
		if err := tx.Model(levelReward).Update("claim_expires_at", expiresAt).Error; err != nil {
			return nil, fmt.Errorf("open level reward: %w", err)
		}
		levelReward.ClaimExpiresAt = &expiresAt
	}

	return levelRewards, nil
}

func levelClaimsPet(tx *gorm.DB, userID uuid.UUID) (models.Pet, error) {
	var userPet models.Pet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&userPet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Pet{}, ErrPetNotFound
	}
	if err != nil {
		return models.Pet{}, fmt.Errorf("lock pet: %w", err)
	}

	return userPet, nil
}

func levelRewardStatus(levelReward models.LevelReward, userLevel int, now time.Time) models.LevelRewardStatus {
	if levelReward.Level > userLevel {
		return models.LevelRewardStatusLocked
	}
	if levelReward.ClaimedAt != nil {
		return models.LevelRewardStatusClaimed
	}
	if levelReward.ClaimExpiresAt == nil || !levelReward.ClaimExpiresAt.After(now) {
		return models.LevelRewardStatusFrozen
	}

	return models.LevelRewardStatusUnopened
}

func levelRewardExpiresAt(levelReward models.LevelReward, status models.LevelRewardStatus) *time.Time {
	if status != models.LevelRewardStatusUnopened && status != models.LevelRewardStatusFrozen {
		return nil
	}
	if levelReward.ClaimExpiresAt == nil {
		return nil
	}

	expiresAt := levelReward.ClaimExpiresAt.UTC()
	return &expiresAt
}
