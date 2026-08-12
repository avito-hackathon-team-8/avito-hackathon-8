package rewards

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGrantRejectsInvalidRewardBeforeDatabaseCall(t *testing.T) {
	t.Parallel()

	service := &Service{now: func() time.Time {
		return time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	}}

	_, err := service.Grant(context.Background(), uuid.New(), Grant{
		Title:     "",
		Category:  models.RewardCategoryAvitoBonus,
		Source:    models.RewardSourceLevel,
		ExpiresAt: time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC),
	})

	if !errors.Is(err, ErrInvalidReward) {
		t.Fatalf("Grant() error = %v, want ErrInvalidReward", err)
	}
}

func TestAllPublishedCategoriesAreValid(t *testing.T) {
	t.Parallel()

	if len(CategoryOrder) != 7 {
		t.Fatalf("len(CategoryOrder) = %d, want 7", len(CategoryOrder))
	}

	for _, category := range CategoryOrder {
		if !validCategory(category) {
			t.Errorf("validCategory(%q) = false", category)
		}
	}
}

func TestRewardSources(t *testing.T) {
	t.Parallel()

	valid := []models.RewardSource{
		models.RewardSourceLevel,
		models.RewardSourceChest,
		models.RewardSourceLeaderboard,
		models.RewardSourceShop,
	}

	for _, source := range valid {
		if !validSource(source) {
			t.Errorf("validSource(%q) = false", source)
		}
	}

	if validSource(models.RewardSource("UNKNOWN")) {
		t.Error("validSource(UNKNOWN) = true")
	}
}

func TestValidGrantOrigin(t *testing.T) {
	t.Parallel()

	levelRewardID := uuid.New()
	chestOpeningID := uuid.New()
	itemType := models.ShopItemTypeCyberBowl

	tests := []struct {
		name  string
		grant Grant
		want  bool
	}{
		{
			name:  "level reward",
			grant: Grant{Source: models.RewardSourceLevel, LevelRewardID: &levelRewardID},
			want:  true,
		},
		{
			name:  "chest reward",
			grant: Grant{Source: models.RewardSourceChest, ChestOpeningID: &chestOpeningID},
			want:  true,
		},
		{
			name:  "leaderboard reward",
			grant: Grant{Source: models.RewardSourceLeaderboard},
			want:  true,
		},
		{
			name:  "shop reward",
			grant: Grant{Source: models.RewardSourceShop, ItemType: &itemType},
			want:  true,
		},
		{
			name:  "chest reward without opening",
			grant: Grant{Source: models.RewardSourceChest},
			want:  false,
		},
		{
			name:  "two origins",
			grant: Grant{Source: models.RewardSourceChest, LevelRewardID: &levelRewardID, ChestOpeningID: &chestOpeningID},
			want:  false,
		},
		{
			name:  "shop reward without item",
			grant: Grant{Source: models.RewardSourceShop},
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := validGrantOrigin(test.grant); got != test.want {
				t.Fatalf("validGrantOrigin(%+v) = %t, want %t", test.grant, got, test.want)
			}
		})
	}
}

func TestRewardStatus(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	service := &Service{now: func() time.Time { return now }}
	redeemedAt := now.Add(-time.Hour)

	tests := []struct {
		name   string
		reward models.Reward
		want   string
	}{
		{
			name:   "active",
			reward: models.Reward{ExpiresAt: now.Add(time.Hour)},
			want:   StatusActive,
		},
		{
			name:   "expired",
			reward: models.Reward{ExpiresAt: now},
			want:   StatusExpired,
		},
		{
			name:   "redeemed takes precedence over expiration",
			reward: models.Reward{ExpiresAt: now.Add(-time.Hour), RedeemedAt: &redeemedAt},
			want:   StatusRedeemed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := service.Status(test.reward); got != test.want {
				t.Fatalf("Status() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRedeemRejectsShopReward(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:rewards-shop-test?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Reward{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	now := time.Date(2030, time.August, 5, 12, 0, 0, 0, time.UTC)
	service := NewService(db, nil)
	service.now = func() time.Time { return now }

	user := models.User{Email: "shop-redeem@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	itemType := models.ShopItemTypeFashionableBowl
	reward := models.Reward{
		UserID: user.ID, Title: "Fashionable bowl", Category: models.RewardCategoryBowl,
		Source: models.RewardSourceShop, ItemType: &itemType, ExpiresAt: now.Add(time.Hour),
	}
	if err := db.Create(&reward).Error; err != nil {
		t.Fatalf("create reward: %v", err)
	}

	if _, err := service.Redeem(context.Background(), user.ID, reward.ID); !errors.Is(err, ErrRewardNotFound) {
		t.Fatalf("Redeem() error = %v, want %v", err, ErrRewardNotFound)
	}
	var stored models.Reward
	if err := db.First(&stored, "id = ?", reward.ID).Error; err != nil {
		t.Fatalf("load reward: %v", err)
	}
	if stored.RedeemedAt != nil {
		t.Fatalf("shop reward redeemedAt = %s, want nil", stored.RedeemedAt)
	}
}
