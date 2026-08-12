package weekly_login

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/petstate"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type failingHappinessProvider struct{}

func (failingHappinessProvider) Get(context.Context, uuid.UUID) (petstate.Snapshot, error) {
	return petstate.Snapshot{}, errors.New("unavailable")
}

func TestCalculateRewardUsesHappinessMultiplier(t *testing.T) {
	tests := []struct {
		happiness float64
		want      int
	}{
		{0, 5}, {50, 10}, {100, 15}, {75, 13}, {-10, 5}, {120, 15},
	}

	for _, test := range tests {
		if got := calculateReward(10, test.happiness); got != test.want {
			t.Errorf("calculateReward(10, %.1f) = %d, want %d", test.happiness, got, test.want)
		}
	}
}

func TestGetFallsBackToBaseRewardsWhenStateIsUnavailable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.UserLogin{}, &models.WeeklyLoginClaim{}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString()), CreatedAt: now.AddDate(0, 0, -1)}

	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.Create(&models.UserLogin{UserID: user.ID, ActivityDate: utcDate(now), CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}

	service := NewService(db, testutil.DailyReportNotifierMock{}, nil, failingHappinessProvider{})
	service.now = func() time.Time { return now }
	week, err := service.Get(context.Background(), user.ID)

	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if week.Claims[4].RewardLeaves != 10 || week.Claims[4].HappinessMultiplier != 1 {
		t.Fatalf("available reward = %+v, want base reward with multiplier 1", week.Claims[4])
	}
}
