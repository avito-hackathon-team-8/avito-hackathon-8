package tasks

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var taskTestNow = time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)

type ensurerFunc func(context.Context, uuid.UUID) error

func (function ensurerFunc) EnsureDailyTasks(ctx context.Context, userID uuid.UUID) error {
	return function(ctx, userID)
}

func TestListUsesCanonicalAssignmentsAndLevelAvailability(t *testing.T) {
	service, _, userID, assignments := testTaskService(t, true)
	items, err := service.List(context.Background(), userID, 1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != TotalDailyTasks {
		t.Fatalf("List() returned %d tasks, want %d", len(items), TotalDailyTasks)
	}
	for index, item := range items {
		if item.ID != assignments[index].ID || item.Slot != index+1 {
			t.Fatalf("task %d = %+v", index, item)
		}
	}
	if items[2].Status != models.LockedTaskStatus || items[3].Status != models.LockedTaskStatus {
		t.Fatalf("high-level tasks are not locked: %+v", items)
	}

	unlocked, err := service.List(context.Background(), userID, 10)
	if err != nil {
		t.Fatalf("List(level 10) error = %v", err)
	}
	if unlocked[2].Status != models.InProgressTaskStatus || unlocked[3].Status != models.InProgressTaskStatus {
		t.Fatalf("high-level tasks were not unlocked: %+v", unlocked)
	}
}

func TestListSynchronouslyEnsuresMissingAssignments(t *testing.T) {
	service, db, userID, _ := testTaskService(t, false)
	called := 0
	service.ensurer = ensurerFunc(func(_ context.Context, gotUserID uuid.UUID) error {
		called++
		if gotUserID != userID {
			t.Fatalf("ensurer user = %s, want %s", gotUserID, userID)
		}
		seedTaskAssignments(t, db, userID)
		return nil
	})

	items, err := service.List(context.Background(), userID, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if called != 1 || len(items) != TotalDailyTasks {
		t.Fatalf("ensure calls = %d, tasks = %d", called, len(items))
	}
}

func TestAutoCompleteFirstTasksIsIdempotent(t *testing.T) {
	service, db, userID, assignments := testTaskService(t, true)

	if err := service.AutoCompleteFirstTasks(context.Background(), userID); err != nil {
		t.Fatalf("AutoCompleteFirstTasks() error = %v", err)
	}

	for index := 0; index < DemoCompletedTaskCount; index++ {
		var assignment models.UserDailyTask
		if err := db.First(&assignment, "id = ?", assignments[index].ID).Error; err != nil {
			t.Fatalf("load assignment %d: %v", index+1, err)
		}
		if assignment.Status != models.CompletedTaskStatus || assignment.CurrentCount != 3 || assignment.CompletedAt == nil {
			t.Fatalf("assignment %d = %+v, want completed at target count", index+1, assignment)
		}
	}

	var thirdBefore models.UserDailyTask
	if err := db.First(&thirdBefore, "id = ?", assignments[2].ID).Error; err != nil {
		t.Fatalf("load third assignment: %v", err)
	}
	if err := service.AutoCompleteFirstTasks(context.Background(), userID); err != nil {
		t.Fatalf("second AutoCompleteFirstTasks() error = %v", err)
	}
	var thirdAfter models.UserDailyTask
	if err := db.First(&thirdAfter, "id = ?", assignments[2].ID).Error; err != nil {
		t.Fatalf("reload third assignment: %v", err)
	}
	if thirdAfter.Status != thirdBefore.Status || thirdAfter.CurrentCount != thirdBefore.CurrentCount || thirdAfter.CompletedAt != thirdBefore.CompletedAt {
		t.Fatalf("third assignment changed: before=%+v after=%+v", thirdBefore, thirdAfter)
	}
}

func TestListReturnsTasksNotReadyWhenEnsureFailsOrCreatesNothing(t *testing.T) {
	service, _, userID, _ := testTaskService(t, false)
	service.ensurer = ensurerFunc(func(context.Context, uuid.UUID) error { return errors.New("puppeteer unavailable") })
	if _, err := service.List(context.Background(), userID, 1); !errors.Is(err, ErrTasksNotReady) {
		t.Fatalf("List() error = %v, want ErrTasksNotReady", err)
	}
	service.ensurer = ensurerFunc(func(context.Context, uuid.UUID) error { return nil })
	if _, err := service.List(context.Background(), userID, 1); !errors.Is(err, ErrTasksNotReady) {
		t.Fatalf("List() after empty ensure error = %v, want ErrTasksNotReady", err)
	}
}

func TestRecordEventsCompletesOnlyAssignedCurrentTask(t *testing.T) {
	service, db, userID, assignments := testTaskService(t, true)
	if err := service.RecordEvents(context.Background(), userID, []Event{{Type: models.ViewListingsTaskType, Count: 2}}, 1); err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}
	var assignment models.UserDailyTask
	if err := db.First(&assignment, "id = ?", assignments[0].ID).Error; err != nil {
		t.Fatalf("load assignment: %v", err)
	}
	if assignment.CurrentCount != 2 || assignment.CompletedAt != nil {
		t.Fatalf("assignment after first event = %+v", assignment)
	}
	if err := service.RecordEvents(context.Background(), userID, []Event{{Type: models.ViewListingsTaskType, Count: int(^uint(0) >> 1)}}, 1); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if err := db.First(&assignment, "id = ?", assignments[0].ID).Error; err != nil {
		t.Fatalf("reload assignment: %v", err)
	}
	if assignment.CurrentCount != 3 || assignment.CompletedAt == nil || assignment.Status != models.CompletedTaskStatus {
		t.Fatalf("completed assignment = %+v", assignment)
	}

	service.now = func() time.Time { return taskTestNow.AddDate(0, 0, 1) }
	if err := service.RecordEvents(context.Background(), userID, []Event{{Type: models.ViewListingsTaskType, Count: 1}}, 1); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("next-day event error = %v, want ErrTaskNotFound", err)
	}
}

func TestRecordEventsRollsBackBatchWhenOneTaskIsLocked(t *testing.T) {
	service, db, userID, assignments := testTaskService(t, true)
	err := service.RecordEvents(context.Background(), userID, []Event{
		{TaskID: assignments[0].ID, Count: 1},
		{TaskID: assignments[2].ID, Count: 1},
	}, 1)
	if !errors.Is(err, ErrTaskLocked) {
		t.Fatalf("RecordEvents() error = %v, want ErrTaskLocked", err)
	}
	var assignment models.UserDailyTask
	if err := db.First(&assignment, "id = ?", assignments[0].ID).Error; err != nil {
		t.Fatalf("load assignment: %v", err)
	}
	if assignment.CurrentCount != 0 {
		t.Fatalf("first assignment count = %d, want rollback to zero", assignment.CurrentCount)
	}
}

func TestRecordEventsRejectsUnknownType(t *testing.T) {
	service, _, userID, _ := testTaskService(t, true)
	if err := service.RecordEvents(context.Background(), userID, []Event{{Type: models.TaskType("UNKNOWN"), Count: 1}}, 10); !errors.Is(err, ErrInvalidTaskType) {
		t.Fatalf("RecordEvents() error = %v, want ErrInvalidTaskType", err)
	}
}

func TestClaimWithRewardIsAtomicAndIdempotent(t *testing.T) {
	service, db, userID, assignments := testTaskService(t, true)
	if err := service.RecordEvents(context.Background(), userID, []Event{{TaskID: assignments[0].ID, Count: 3}}, 1); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	rewardErr := errors.New("reward failed")
	if _, err := service.ClaimWithReward(context.Background(), userID, assignments[0].ID, 1, func(*gorm.DB, int) error { return rewardErr }); !errors.Is(err, rewardErr) {
		t.Fatalf("ClaimWithReward() error = %v, want reward error", err)
	}
	var assignment models.UserDailyTask
	if err := db.First(&assignment, "id = ?", assignments[0].ID).Error; err != nil {
		t.Fatalf("load assignment: %v", err)
	}
	if assignment.ClaimedAt != nil {
		t.Fatal("failed reward committed claimedAt")
	}

	result, err := service.ClaimWithReward(context.Background(), userID, assignments[0].ID, 1, func(_ *gorm.DB, amount int) error {
		if amount != 45 {
			t.Fatalf("reward amount = %d, want 45", amount)
		}
		return nil
	})
	if err != nil || result.Status != models.ClaimedTaskStatus {
		t.Fatalf("successful claim = %+v, error = %v", result, err)
	}
	if _, err := service.ClaimWithReward(context.Background(), userID, assignments[0].ID, 1, nil); !errors.Is(err, ErrRewardAlreadyClaimed) {
		t.Fatalf("second claim error = %v, want ErrRewardAlreadyClaimed", err)
	}
}

func TestProgressCountsCompletedAndClaimed(t *testing.T) {
	service, db, userID, assignments := testTaskService(t, true)
	now := taskTestNow
	if err := db.Model(&models.UserDailyTask{}).Where("id = ?", assignments[0].ID).Updates(map[string]any{"status": models.CompletedTaskStatus, "completed_at": now}).Error; err != nil {
		t.Fatalf("complete assignment: %v", err)
	}
	if err := db.Model(&models.UserDailyTask{}).Where("id = ?", assignments[1].ID).Updates(map[string]any{"status": models.ClaimedTaskStatus, "completed_at": now, "claimed_at": now}).Error; err != nil {
		t.Fatalf("claim assignment: %v", err)
	}
	progress, err := service.Progress(context.Background(), userID, 10)
	if err != nil {
		t.Fatalf("Progress() error = %v", err)
	}
	if progress.CompletedCount != 2 || progress.TotalCount != TotalDailyTasks {
		t.Fatalf("progress = %+v", progress)
	}
}

func testTaskService(t *testing.T, withAssignments bool) (*Service, *gorm.DB, uuid.UUID, []models.UserDailyTask) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.DailyTaskDefinition{}, &models.UserDailyTask{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	service := NewService(db, testutil.DailyReportNotifierMock{})
	service.now = func() time.Time { return taskTestNow }
	userID := uuid.New()
	var assignments []models.UserDailyTask
	if withAssignments {
		assignments = seedTaskAssignments(t, db, userID)
	}
	return service, db, userID, assignments
}

func seedTaskAssignments(t *testing.T, db *gorm.DB, userID uuid.UUID) []models.UserDailyTask {
	t.Helper()
	types := []models.TaskType{models.ViewListingsTaskType, models.AddToFavoritesTaskType, models.PublishListingTaskType, models.BoostListingTaskType}
	assignments := make([]models.UserDailyTask, 0, TotalDailyTasks)
	for index, taskType := range types {
		level := 1
		switch index {
		case 2:
			level = 5
		case 3:
			level = 10
		}
		definition := models.DailyTaskDefinition{Code: fmt.Sprintf("task-%d", index+1), Title: fmt.Sprintf("Task %d", index+1), Slot: index + 1, Type: taskType, TargetCount: 3, Reward: 45, UnlockLevel: level, Categories: "[]", Active: true}
		if err := db.Create(&definition).Error; err != nil {
			t.Fatalf("create definition: %v", err)
		}
		status := models.InProgressTaskStatus
		if level > 1 {
			status = models.LockedTaskStatus
		}
		assignment := models.UserDailyTask{UserID: userID, TaskDefinitionID: definition.ID, Day: utcDate(taskTestNow), Status: status, ExpiresAt: utcDate(taskTestNow).AddDate(0, 0, 1)}
		if err := db.Create(&assignment).Error; err != nil {
			t.Fatalf("create assignment: %v", err)
		}
		assignments = append(assignments, assignment)
	}
	return assignments
}
