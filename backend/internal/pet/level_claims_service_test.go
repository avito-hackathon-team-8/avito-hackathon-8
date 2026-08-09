package pet

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/reward_catalog"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGetLevelsCreatesTrackAndOpensReachedLevels(t *testing.T) {
	service, _, user, now := levelClaimsTestService(t, 2)

	levels, err := service.GetLevels(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetLevels() error = %v", err)
	}
	if len(levels) != MaxPetLevel {
		t.Fatalf("GetLevels() returned %d levels, want %d", len(levels), MaxPetLevel)
	}

	for _, level := range levels {
		if level.Level <= 2 {
			if level.Status != models.LevelRewardStatusUnopened {
				t.Errorf("level %d status = %s, want UNOPENED", level.Level, level.Status)
			}
			if level.ExpiresAt == nil || !level.ExpiresAt.Equal(now.Add(LevelRewardClaimWindow)) {
				t.Errorf("level %d expiresAt = %v, want %v", level.Level, level.ExpiresAt, now.Add(LevelRewardClaimWindow))
			}
			continue
		}
		if level.Status != models.LevelRewardStatusLocked || level.ExpiresAt != nil {
			t.Errorf("level %d = %+v, want LOCKED without expiresAt", level.Level, level)
		}
	}
}

func TestClaimIssuesOneRewardAndMarksLevelClaimed(t *testing.T) {
	service, db, user, _ := levelClaimsTestService(t, 1)
	levels, err := service.GetLevels(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetLevels() error = %v", err)
	}

	result, err := service.Claim(context.Background(), user.ID, levels[0].Reward.ID)
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if result.Level != 1 || result.Status != models.LevelRewardStatusClaimed {
		t.Fatalf("Claim() result = %+v, want level 1 CLAIMED", result)
	}
	if result.Reward.Source != models.RewardSourceLevel || result.Reward.LevelRewardID == nil || *result.Reward.LevelRewardID != levels[0].Reward.ID {
		t.Fatalf("issued reward = %+v, want LEVEL source linked to level reward", result.Reward)
	}

	var storedLevelReward models.LevelReward
	if err := db.Where("id = ?", levels[0].Reward.ID).First(&storedLevelReward).Error; err != nil {
		t.Fatalf("load level reward: %v", err)
	}
	if storedLevelReward.ClaimedAt == nil {
		t.Fatal("Claim() did not persist claimedAt")
	}

	var issuedCount int64
	if err := db.Model(&models.Reward{}).Where("level_reward_id = ?", levels[0].Reward.ID).Count(&issuedCount).Error; err != nil {
		t.Fatalf("count issued rewards: %v", err)
	}
	if issuedCount != 1 {
		t.Fatalf("issued rewards = %d, want 1", issuedCount)
	}

	if _, err := service.Claim(context.Background(), user.ID, levels[0].Reward.ID); !errors.Is(err, ErrLevelRewardAlreadyClaimed) {
		t.Fatalf("second Claim() error = %v, want %v", err, ErrLevelRewardAlreadyClaimed)
	}
}

func TestClaimRejectsLockedFrozenAndForeignRewards(t *testing.T) {
	service, db, user, now := levelClaimsTestService(t, 1)
	levels, err := service.GetLevels(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetLevels() error = %v", err)
	}

	if _, err := service.Claim(context.Background(), user.ID, levels[1].Reward.ID); !errors.Is(err, ErrLevelRewardLocked) {
		t.Fatalf("Claim(locked) error = %v, want %v", err, ErrLevelRewardLocked)
	}

	if err := db.Model(&models.Pet{}).Where("user_id = ?", user.ID).Update("level", 2).Error; err != nil {
		t.Fatalf("raise pet level: %v", err)
	}
	if _, err := service.GetLevels(context.Background(), user.ID); err != nil {
		t.Fatalf("GetLevels() after level up error = %v", err)
	}
	expiredAt := now.Add(-time.Second)
	if err := db.Model(&models.LevelReward{}).Where("id = ?", levels[1].Reward.ID).Update("claim_expires_at", expiredAt).Error; err != nil {
		t.Fatalf("expire level reward: %v", err)
	}
	if _, err := service.Claim(context.Background(), user.ID, levels[1].Reward.ID); !errors.Is(err, ErrLevelRewardFrozen) {
		t.Fatalf("Claim(frozen) error = %v, want %v", err, ErrLevelRewardFrozen)
	}

	otherUser := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}
	if err := db.Create(&otherUser).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}
	if err := db.Create(&models.Pet{UserID: otherUser.ID, Level: 1}).Error; err != nil {
		t.Fatalf("create other pet: %v", err)
	}
	if _, err := service.Claim(context.Background(), otherUser.ID, levels[0].Reward.ID); !errors.Is(err, ErrLevelRewardNotFound) {
		t.Fatalf("Claim(foreign reward) error = %v, want %v", err, ErrLevelRewardNotFound)
	}
}

func TestClaimRollsBackWhenRewardCannotBeIssued(t *testing.T) {
	service, db, user, now := levelClaimsTestService(t, 1)
	levels, err := service.GetLevels(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetLevels() error = %v", err)
	}

	existing := models.Reward{
		UserID:        user.ID,
		LevelRewardID: &levels[0].Reward.ID,
		Title:         "existing",
		Category:      models.RewardCategoryAvitoBonus,
		Source:        models.RewardSourceLevel,
		ExpiresAt:     now.Add(time.Hour),
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("create conflicting reward: %v", err)
	}

	if _, err := service.Claim(context.Background(), user.ID, levels[0].Reward.ID); err == nil {
		t.Fatal("Claim() error = nil, want unique constraint error")
	}

	var storedLevelReward models.LevelReward
	if err := db.Where("id = ?", levels[0].Reward.ID).First(&storedLevelReward).Error; err != nil {
		t.Fatalf("load level reward: %v", err)
	}
	if storedLevelReward.ClaimedAt != nil {
		t.Fatal("Claim() persisted claimedAt when reward creation failed")
	}
}

func TestLevelUpStartsClaimWindowInPetTransaction(t *testing.T) {
	levelClaims, db, user, now := levelClaimsTestService(t, 1)
	petService := NewService(db)
	petService.SetLevelClaimsService(levelClaims)

	if _, err := petService.AddLeaves(context.Background(), user.ID, 100); err != nil {
		t.Fatalf("AddLeaves() error = %v", err)
	}

	var levelReward models.LevelReward
	if err := db.Where("user_id = ? AND level = ?", user.ID, 2).First(&levelReward).Error; err != nil {
		t.Fatalf("load level 2 reward: %v", err)
	}
	if levelReward.ClaimExpiresAt == nil || !levelReward.ClaimExpiresAt.Equal(now.Add(LevelRewardClaimWindow)) {
		t.Fatalf("level 2 claim deadline = %v, want %v", levelReward.ClaimExpiresAt, now.Add(LevelRewardClaimWindow))
	}
}

func levelClaimsTestService(t *testing.T, level int) (*LevelClaimsService, *gorm.DB, models.User, time.Time) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Pet{},
		&models.LevelReward{},
		&models.Reward{},
		&models.LeafTransaction{},
		&models.UserGameState{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.Pet{UserID: user.ID, Level: level}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	notifier := testutil.DailyReportNotifierMock{}
	rewardService := rewards.NewService(db, notifier)

	service := NewLevelClaimsService(db, notifier, rewardService, testLevelRewardDefinitions())
	service.now = func() time.Time { return now }

	return service, db, user, now
}

func testLevelRewardDefinitions() []LevelRewardDefinition {
	definitions := make([]LevelRewardDefinition, 0, MaxPetLevel)
	for level := 1; level <= MaxPetLevel; level++ {
		definitions = append(definitions, reward_catalog.LevelRewardDefinition{
			Level:       level,
			Title:       fmt.Sprintf("Награда за уровень %d", level),
			Description: fmt.Sprintf("Награда за уровень %d", level),
			Category:    models.RewardCategoryAvitoBonus,
		})
	}

	return definitions
}
