package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRewardOriginConstraint(t *testing.T) {
	db := rewardTestDatabase(t)
	now := time.Now().UTC()
	user := User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	levelReward := LevelReward{
		UserID:      user.ID,
		Level:       1,
		Title:       "100 бонусов Авито",
		Description: "100 бонусов Авито",
		Category:    RewardCategoryAvitoBonus,
	}
	if err := db.Create(&levelReward).Error; err != nil {
		t.Fatalf("create level reward: %v", err)
	}

	opening := ChestOpening{UserID: user.ID, LeavesSpent: ChestOpeningLeavesCost, OpenedAt: now}
	if err := db.Create(&opening).Error; err != nil {
		t.Fatalf("create chest opening: %v", err)
	}

	tests := []struct {
		name   string
		reward Reward
		valid  bool
	}{
		{
			name:   "level reward",
			reward: Reward{UserID: user.ID, LevelRewardID: &levelReward.ID, Title: "100 бонусов Авито", Category: RewardCategoryAvitoBonus, Source: RewardSourceLevel, ExpiresAt: now.Add(time.Hour)},
			valid:  true,
		},
		{
			name:   "chest reward",
			reward: Reward{UserID: user.ID, ChestOpeningID: &opening.ID, Title: "1000 бонусов Авито", Category: RewardCategoryAvitoBonus, Source: RewardSourceChest, ExpiresAt: now.Add(time.Hour)},
			valid:  true,
		},
		{
			name:   "leaderboard reward",
			reward: Reward{UserID: user.ID, Title: "1000 бонусов Авито", Category: RewardCategoryAvitoBonus, Source: RewardSourceLeaderboard, ExpiresAt: now.Add(time.Hour)},
			valid:  true,
		},
		{
			name:   "chest reward without opening",
			reward: Reward{UserID: user.ID, Title: "1000 бонусов Авито", Category: RewardCategoryAvitoBonus, Source: RewardSourceChest, ExpiresAt: now.Add(time.Hour)},
		},
		{
			name:   "two reward origins",
			reward: Reward{UserID: user.ID, LevelRewardID: &levelReward.ID, ChestOpeningID: &opening.ID, Title: "1000 бонусов Авито", Category: RewardCategoryAvitoBonus, Source: RewardSourceChest, ExpiresAt: now.Add(time.Hour)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := db.Create(&test.reward).Error
			if test.valid && err != nil {
				t.Fatalf("create reward: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("create reward error = nil, want check constraint violation")
			}
		})
	}
}

func rewardTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &LevelReward{}, &ChestOpening{}, &Reward{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return db
}
