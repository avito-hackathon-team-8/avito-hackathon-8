package weekly_login

import (
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
)

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

	week := buildCurrentWeek(user, claims, today, ActivityStatusActive)

	if week.ClaimedDaysCount != 1 {
		t.Fatalf("ClaimedDaysCount = %d, want 1", week.ClaimedDaysCount)
	}

	if len(week.Claims) != 7 {
		t.Fatalf("len(Claims) = %d, want 7", len(week.Claims))
	}

	wantStatuses := []DayStatus{
		DayStatusClaimed,
		DayStatusMissed,
		DayStatusAvailable,
		DayStatusFuture,
		DayStatusFuture,
		DayStatusFuture,
		DayStatusFuture,
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

	week := buildCurrentWeek(user, nil, today, ActivityStatusActive)

	if week.Claims[0].Status != DayStatusBeforeRegistration ||
		week.Claims[1].Status != DayStatusBeforeRegistration {
		t.Fatalf("days before registration = (%q, %q), want BEFORE_REGISTRATION", week.Claims[0].Status, week.Claims[1].Status)
	}

	if week.Claims[2].Status != DayStatusAvailable || week.Claims[2].RewardLeaves != 10 {
		t.Fatalf("registration day = (%q, %d), want (AVAILABLE, 10)", week.Claims[2].Status, week.Claims[2].RewardLeaves)
	}
}

func TestBuildCurrentWeekMarksInactiveTodayUnconfirmed(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	user := models.User{CreatedAt: time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)}

	week := buildCurrentWeek(user, nil, today, ActivityStatusInactive)

	if week.Claims[2].Status != DayStatusUnconfirmed || week.Claims[2].RewardLeaves != 0 {
		t.Fatalf("inactive today = (%q, %d), want (UNCONFIRMED, 0)", week.Claims[2].Status, week.Claims[2].RewardLeaves)
	}

	if week.Claims[3].Status != DayStatusFuture || week.Claims[3].RewardLeaves != 10 {
		t.Fatalf("next day = (%q, %d), want (FUTURE, 10)", week.Claims[3].Status, week.Claims[3].RewardLeaves)
	}
}
