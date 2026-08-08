package rewards

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
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

	if len(CategoryOrder) != 5 {
		t.Fatalf("len(CategoryOrder) = %d, want 5", len(CategoryOrder))
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
			name:  "chest reward without opening",
			grant: Grant{Source: models.RewardSourceChest},
			want:  false,
		},
		{
			name:  "two origins",
			grant: Grant{Source: models.RewardSourceChest, LevelRewardID: &levelRewardID, ChestOpeningID: &chestOpeningID},
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
