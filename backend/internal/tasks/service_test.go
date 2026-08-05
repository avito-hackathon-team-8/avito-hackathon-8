package tasks

import (
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
)

func TestStatusFor(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	claimedAt := now.Add(time.Hour)

	tests := []struct {
		name      string
		task      models.Task
		progress  *models.UserTaskProgress
		userLevel int
		want      models.TaskStatus
	}{
		{
			name:      "locked",
			task:      models.Task{RequiredLevel: 5, TargetCount: 3},
			userLevel: 1,
			want:      models.LockedTaskStatus,
		},
		{
			name:      "in progress without progress record",
			task:      models.Task{RequiredLevel: 1, TargetCount: 3},
			userLevel: 1,
			want:      models.InProgressTaskStatus,
		},
		{
			name:      "claimed",
			task:      models.Task{RequiredLevel: 1, TargetCount: 3},
			progress:  &models.UserTaskProgress{CurrentCount: 3, CompletedAt: &now, ClaimedAt: &claimedAt},
			userLevel: 1,
			want:      models.ClaimedTaskStatus,
		},
		{
			name:      "completed by timestamp",
			task:      models.Task{RequiredLevel: 1, TargetCount: 3},
			progress:  &models.UserTaskProgress{CurrentCount: 2, CompletedAt: &now},
			userLevel: 1,
			want:      models.CompletedTaskStatus,
		},
		{
			name:      "completed by count",
			task:      models.Task{RequiredLevel: 1, TargetCount: 3},
			progress:  &models.UserTaskProgress{CurrentCount: 3},
			userLevel: 1,
			want:      models.CompletedTaskStatus,
		},
		{
			name:      "in progress",
			task:      models.Task{RequiredLevel: 1, TargetCount: 3},
			progress:  &models.UserTaskProgress{CurrentCount: 2},
			userLevel: 1,
			want:      models.InProgressTaskStatus,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := StatusFor(test.task, test.progress, test.userLevel); got != test.want {
				t.Fatalf("StatusFor() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTodayUsesUTCMidnight(t *testing.T) {
	t.Parallel()

	service := &Service{now: func() time.Time {
		return time.Date(2026, time.August, 5, 23, 30, 0, 0, time.FixedZone("MSK", 3*60*60))
	}}

	want := time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC)

	if got := service.today(); !got.Equal(want) {
		t.Fatalf("today() = %v, want %v", got, want)
	}
}

func TestDailyTaskDefinitions(t *testing.T) {
	t.Parallel()

	if len(dailyTaskDefinitions) != TotalDailyTasks {
		t.Fatalf("len(dailyTaskDefinitions) = %d, want %d", len(dailyTaskDefinitions), TotalDailyTasks)
	}

	slots := make(map[int]bool, TotalDailyTasks)

	for _, definition := range dailyTaskDefinitions {
		if definition.Slot < 1 || definition.Slot > TotalDailyTasks {
			t.Errorf("definition slot = %d, want 1..%d", definition.Slot, TotalDailyTasks)
		}
		if slots[definition.Slot] {
			t.Errorf("duplicate definition slot %d", definition.Slot)
		}
		slots[definition.Slot] = true

		if definition.TargetCount < 1 {
			t.Errorf("definition %d target count = %d, want positive", definition.Slot, definition.TargetCount)
		}
		if definition.RewardLeaves < 1 {
			t.Errorf("definition %d reward leaves = %d, want positive", definition.Slot, definition.RewardLeaves)
		}
		if definition.RequiredLevel < 1 {
			t.Errorf("definition %d required level = %d, want positive", definition.Slot, definition.RequiredLevel)
		}
		if definition.Description == "" {
			t.Errorf("definition %d description is empty", definition.Slot)
		}
	}
}
