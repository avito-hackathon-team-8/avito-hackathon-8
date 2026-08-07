package weekly_login

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
)

type activityProviderStub struct {
	day ActivityDay
	err error
}

func (stub activityProviderStub) Add(context.Context, uuid.UUID, []ActivityDay) error {
	return nil
}

func (stub activityProviderStub) Get(context.Context, uuid.UUID, time.Time) (ActivityDay, error) {
	return stub.day, stub.err
}

func (stub activityProviderStub) GetRange(context.Context, uuid.UUID, time.Time, time.Time) ([]ActivityDay, error) {
	return nil, nil
}

func TestUTCDate(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.August, 5, 2, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	want := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)

	if got := utcDate(date); !got.Equal(want) {
		t.Fatalf("utcDate(%s) = %s, want %s", date, got, want)
	}
}

func TestUTCWeekBounds(t *testing.T) {
	t.Parallel()

	wantStart := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)

	for day := 3; day <= 9; day++ {
		date := time.Date(2026, time.August, day, 18, 45, 0, 0, time.UTC)
		start, end := utcWeekBounds(date)

		if !start.Equal(wantStart) || !end.Equal(wantEnd) {
			t.Errorf("utcWeekBounds(%s) = (%s, %s), want (%s, %s)", date, start, end, wantStart, wantEnd)
		}
	}
}

func TestWeeklyRewardByIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		index int
		want  weeklyReward
	}{
		{index: 0, want: weeklyRewardFirst},
		{index: 1, want: weeklyRewardSecond},
		{index: 2, want: weeklyRewardThird},
		{index: 3, want: weeklyRewardFourth},
		{index: 4, want: weeklyRewardFifth},
		{index: 5, want: weeklyRewardSixth},
		{index: 6, want: weeklyRewardSeventh},
		{index: -1, want: 0},
		{index: 7, want: 0},
	}

	for _, test := range tests {
		if got := weeklyRewardByIndex(test.index); got != test.want {
			t.Errorf("weeklyRewardByIndex(%d) = %d, want %d", test.index, got, test.want)
		}
	}
}

func TestBuildCurrentWeek(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	claimID := uuid.New()
	user := models.User{CreatedAt: time.Date(2026, time.August, 3, 10, 0, 0, 0, time.UTC)}
	claims := []models.WeeklyLoginClaim{
		{
			ID:           claimID,
			ClaimDate:    time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
			RewardLeaves: 10,
		},
	}

	week := buildCurrentWeek(user, claims, today, false)

	if week.ClaimedDaysCount != 1 {
		t.Fatalf("ClaimedDaysCount = %d, want 1", week.ClaimedDaysCount)
	}

	if len(week.Claims) != 7 {
		t.Fatalf("len(Claims) = %d, want 7", len(week.Claims))
	}

	wantStatuses := []models.DayStatus{
		models.DayStatusClaimed,
		models.DayStatusMissed,
		models.DayStatusAvailable,
		models.DayStatusFuture,
		models.DayStatusFuture,
		models.DayStatusFuture,
		models.DayStatusFuture,
	}
	wantRewards := []int{10, 0, 20, 30, 40, 50, 60}

	for index, day := range week.Claims {
		if day.Weekday != index+1 {
			t.Errorf("Claims[%d].Weekday = %d, want %d", index, day.Weekday, index+1)
		}

		if day.Status != wantStatuses[index] {
			t.Errorf("Claims[%d].Status = %q, want %q", index, day.Status, wantStatuses[index])
		}

		if day.RewardLeaves != wantRewards[index] {
			t.Errorf("Claims[%d].RewardLeaves = %d, want %d", index, day.RewardLeaves, wantRewards[index])
		}
	}

	if week.Claims[0].ClaimID == nil || *week.Claims[0].ClaimID != claimID {
		t.Fatalf("claimed day ClaimID = %v, want %s", week.Claims[0].ClaimID, claimID)
	}
}

func TestBuildCurrentWeekMarksDaysBeforeRegistration(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	user := models.User{CreatedAt: today}

	week := buildCurrentWeek(user, nil, today, false)

	if week.Claims[0].Status != models.DayStatusBeforeRegistration ||
		week.Claims[1].Status != models.DayStatusBeforeRegistration {
		t.Fatalf("days before registration = (%q, %q), want BEFORE_REGISTRATION", week.Claims[0].Status, week.Claims[1].Status)
	}

	if week.Claims[2].Status != models.DayStatusAvailable || week.Claims[2].RewardLeaves != 10 {
		t.Fatalf("registration day = (%q, %d), want (AVAILABLE, 10)", week.Claims[2].Status, week.Claims[2].RewardLeaves)
	}
}

func TestBuildCurrentWeekMarksInactiveTodayUnconfirmed(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	user := models.User{CreatedAt: time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)}

	week := buildCurrentWeek(user, nil, today, true)

	if week.Claims[2].Status != models.DayStatusUnconfirmed || week.Claims[2].RewardLeaves != 0 {
		t.Fatalf("inactive today = (%q, %d), want (UNCONFIRMED, 0)", week.Claims[2].Status, week.Claims[2].RewardLeaves)
	}

	if week.Claims[3].Status != models.DayStatusFuture || week.Claims[3].RewardLeaves != 10 {
		t.Fatalf("next day = (%q, %d), want (FUTURE, 10)", week.Claims[3].Status, week.Claims[3].RewardLeaves)
	}
}

func TestActivityInactive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider ActivityProvider
		want     bool
	}{
		{name: "active", provider: activityProviderStub{day: ActivityDay{Active: true}}, want: false},
		{name: "inactive", provider: activityProviderStub{day: ActivityDay{Active: false}}, want: true},
		{name: "provider error is fail open", provider: activityProviderStub{err: errors.New("unavailable")}, want: false},
		{name: "provider is not configured", provider: nil, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &Service{activity: test.provider}
			got := service.activityInactive(t.Context(), uuid.New(), time.Now())

			if got != test.want {
				t.Fatalf("activityInactive() = %t, want %t", got, test.want)
			}
		})
	}
}
