package tasks

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func TestRecordEventsAddsProgressAndCompletesTask(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	err := service.RecordEvents(context.Background(), userID, []Event{
		{Type: models.OpenNotificationsTaskType, Count: 2},
	}, 1)
	if err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}

	task := testTask(t, service, models.OpenNotificationsTaskType)
	progress := testProgress(t, service, userID, task.ID)
	if progress.CurrentCount != 2 || progress.CompletedAt != nil {
		t.Fatalf("progress after first event = %+v, want count 2 without completion", progress)
	}

	err = service.RecordEvents(context.Background(), userID, []Event{
		{Type: models.OpenNotificationsTaskType, Count: 10},
	}, 1)
	if err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}

	progress = testProgress(t, service, userID, task.ID)
	if progress.CurrentCount != task.TargetCount || progress.CompletedAt == nil {
		t.Fatalf("progress after completing event = %+v, want count %d and completed time", progress, task.TargetCount)
	}
}

func TestRecordEventsRollsBackBatchForLockedTask(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	err := service.RecordEvents(context.Background(), userID, []Event{
		{Type: models.AddToFavoritesTaskType, Count: 1},
		{Type: models.PublishListingTaskType, Count: 1},
	}, 1)
	if !errors.Is(err, ErrTaskLocked) {
		t.Fatalf("RecordEvents() error = %v, want ErrTaskLocked", err)
	}

	task := testTask(t, service, models.AddToFavoritesTaskType)
	var count int64
	if err := service.db.Model(&models.UserTaskProgress{}).
		Where("user_id = ? AND task_id = ?", userID, task.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count progress: %v", err)
	}
	if count != 0 {
		t.Fatalf("progress records = %d, want 0 after rolled back batch", count)
	}
}

func TestClaimMarksCompletedTaskAsClaimed(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	err := service.RecordEvents(context.Background(), userID, []Event{
		{Type: models.OpenNotificationsTaskType, Count: 5},
	}, 1)
	if err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}

	task := testTask(t, service, models.OpenNotificationsTaskType)
	result, err := service.Claim(context.Background(), userID, task.ID, 1)
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if result.TaskID != task.ID || result.RewardLeaves != task.RewardLeaves || result.Status != models.ClaimedTaskStatus {
		t.Fatalf("Claim() result = %+v, want task %s, reward %d and CLAIMED", result, task.ID, task.RewardLeaves)
	}

	progress := testProgress(t, service, userID, task.ID)
	if progress.ClaimedAt == nil {
		t.Fatal("Claim() did not set claimed time")
	}

	_, err = service.Claim(context.Background(), userID, task.ID, 1)
	if !errors.Is(err, ErrRewardAlreadyClaimed) {
		t.Fatalf("second Claim() error = %v, want ErrRewardAlreadyClaimed", err)
	}
}

func TestClaimRejectsIncompleteAndLockedTasks(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	if _, err := service.List(context.Background(), userID, 10); err != nil {
		t.Fatalf("List() error = %v", err)
	}

	openNotifications := testTask(t, service, models.OpenNotificationsTaskType)
	_, err := service.Claim(context.Background(), userID, openNotifications.ID, 1)
	if !errors.Is(err, ErrTaskNotCompleted) {
		t.Fatalf("Claim() error = %v, want ErrTaskNotCompleted", err)
	}

	publishListing := testTask(t, service, models.PublishListingTaskType)
	_, err = service.Claim(context.Background(), userID, publishListing.ID, 1)
	if !errors.Is(err, ErrTaskLocked) {
		t.Fatalf("Claim() error = %v, want ErrTaskLocked", err)
	}
}

func testService(t *testing.T) *Service {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Task{}, &models.UserTaskProgress{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return &Service{
		db: db,
		now: func() time.Time {
			return time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
		},
	}
}

func testTask(t *testing.T, service *Service, taskType models.TaskType) models.Task {
	t.Helper()

	var task models.Task
	if err := service.db.Where("date = ? AND type = ?", service.today(), taskType).First(&task).Error; err != nil {
		t.Fatalf("find task %s: %v", taskType, err)
	}

	return task
}

func testProgress(t *testing.T, service *Service, userID, taskID uuid.UUID) models.UserTaskProgress {
	t.Helper()

	var progress models.UserTaskProgress
	if err := service.db.Where("user_id = ? AND task_id = ?", userID, taskID).First(&progress).Error; err != nil {
		t.Fatalf("find task progress: %v", err)
	}

	return progress
}
