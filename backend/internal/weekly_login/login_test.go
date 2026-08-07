package weekly_login

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuildActivityDaysReturnsInclusiveRange(t *testing.T) {
	t.Parallel()

	dateFrom := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, time.August, 7, 0, 0, 0, 0, time.UTC)
	activeDates := map[string]struct{}{
		"2026-08-05": {},
		"2026-08-06": {},
	}

	days := buildActivityDays(dateFrom, dateTo, activeDates)
	wantActive := []bool{false, false, true, true, false}

	if len(days) != len(wantActive) {
		t.Fatalf("len(days) = %d, want %d", len(days), len(wantActive))
	}

	for index, day := range days {
		wantDate := dateFrom.AddDate(0, 0, index)

		if !day.Date.Equal(wantDate) {
			t.Errorf("days[%d].Date = %s, want %s", index, day.Date, wantDate)
		}

		if day.Active != wantActive[index] {
			t.Errorf("days[%d].Active = %t, want %t", index, day.Active, wantActive[index])
		}
	}
}

func TestActivityServiceRejectsInvalidRange(t *testing.T) {
	t.Parallel()

	service := &ActivityService{}
	dateFrom := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	if _, err := service.GetRange(t.Context(), uuid.Nil, dateFrom, dateFrom.AddDate(0, 0, 7)); !errors.Is(err, ErrInvalidActivityRange) {
		t.Fatalf("GetRange() error = %v, want ErrInvalidActivityRange", err)
	}
}
