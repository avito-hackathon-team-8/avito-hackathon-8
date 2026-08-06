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

	definitions, err := LoadDefaultDefinitions()
	if err != nil {
		t.Fatalf("LoadDefaultDefinitions() error = %v", err)
	}

	slots := make(map[int]int, TotalDailyTasks)
	types := make(map[models.TaskType]bool)

	for _, definition := range definitions {
		if definition.Slot < 1 || definition.Slot > TotalDailyTasks {
			t.Errorf("definition slot = %d, want 1..%d", definition.Slot, TotalDailyTasks)
		}
		slots[definition.Slot]++
		types[definition.Type] = true

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

	for slot, count := range slots {
		if count == 0 {
			t.Errorf("slot %d has no definitions", slot)
		}
	}
	for _, taskType := range knownTaskTypes {
		if !types[taskType] {
			t.Errorf("type %s has no definitions", taskType)
		}
	}
}

func TestRecordEventsAddsProgressAndCompletesTask(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	err := service.RecordEvents(context.Background(), userID, []Event{
		{Type: models.ViewListingsTaskType, Count: 2},
	}, 1)
	if err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}

	task := testTask(t, service, models.ViewListingsTaskType)
	progress := testProgress(t, service, userID, task.ID)
	if progress.CurrentCount != 2 || progress.CompletedAt != nil {
		t.Fatalf("progress after first event = %+v, want count 2 without completion", progress)
	}

	err = service.RecordEvents(context.Background(), userID, []Event{
		{Type: models.ViewListingsTaskType, Count: 10},
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
	lockedDefinition := service.definitionsForDate(service.today())[2]

	err := service.RecordEvents(context.Background(), userID, []Event{
		{Type: models.AddToFavoritesTaskType, Count: 1},
		{Type: lockedDefinition.Type, Count: 1},
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

func TestRecordEventsRejectsUnknownTaskType(t *testing.T) {
	service := testService(t)
	err := service.RecordEvents(context.Background(), uuid.New(), []Event{{Type: models.TaskType("UNKNOWN"), Count: 1}}, 1)
	if !errors.Is(err, ErrInvalidTaskType) {
		t.Fatalf("RecordEvents() error = %v, want ErrInvalidTaskType", err)
	}
}

func TestClaimMarksCompletedTaskAsClaimed(t *testing.T) {
	service := testService(t)
	userID := uuid.New()
	if _, err := service.List(context.Background(), userID, 1); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	task := testTaskBySlot(t, service, 1)

	err := service.RecordEvents(context.Background(), userID, []Event{
		{Type: task.Type, Count: task.TargetCount},
	}, 1)
	if err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}

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

func TestClaimWithRewardRollsBackWhenRewardApplicationFails(t *testing.T) {
	service := testService(t)
	userID := uuid.New()
	if _, err := service.List(context.Background(), userID, 1); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	task := testTaskBySlot(t, service, 1)
	if err := service.RecordEvents(context.Background(), userID, []Event{{Type: task.Type, Count: task.TargetCount}}, 1); err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}

	rewardErr := errors.New("reward application failed")
	if _, err := service.ClaimWithReward(context.Background(), userID, task.ID, 1, func(_ *gorm.DB, _ int) error {
		return rewardErr
	}); !errors.Is(err, rewardErr) {
		t.Fatalf("ClaimWithReward() error = %v, want %v", err, rewardErr)
	}

	progress := testProgress(t, service, userID, task.ID)
	if progress.ClaimedAt != nil {
		t.Fatal("ClaimWithReward() committed claimed time after reward failure")
	}
}

func TestClaimRejectsIncompleteAndLockedTasks(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	if _, err := service.List(context.Background(), userID, 10); err != nil {
		t.Fatalf("List() error = %v", err)
	}

	viewListings := testTaskBySlot(t, service, 1)
	_, err := service.Claim(context.Background(), userID, viewListings.ID, 1)
	if !errors.Is(err, ErrTaskNotCompleted) {
		t.Fatalf("Claim() error = %v, want ErrTaskNotCompleted", err)
	}

	lockedTask := testTaskBySlot(t, service, 3)
	_, err = service.Claim(context.Background(), userID, lockedTask.ID, 1)
	if !errors.Is(err, ErrTaskLocked) {
		t.Fatalf("Claim() error = %v, want ErrTaskLocked", err)
	}
}

func TestListReturnsFourSlotsAndProgress(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	items, err := service.List(context.Background(), userID, 1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != TotalDailyTasks {
		t.Fatalf("List() returned %d tasks, want %d", len(items), TotalDailyTasks)
	}
	for index, item := range items {
		wantSlot := index + 1
		if item.Slot != wantSlot {
			t.Errorf("task %d slot = %d, want %d", index, item.Slot, wantSlot)
		}
		if item.CurrentCount != 0 || item.Status == models.ClaimedTaskStatus {
			t.Errorf("task %d before events = %+v", index, item)
		}
	}

	firstTask := testTaskBySlot(t, service, 1)
	if err := service.RecordEvent(context.Background(), userID, firstTask.Type, 1); err != nil {
		t.Fatalf("RecordEvent() error = %v", err)
	}

	items, err = service.List(context.Background(), userID, 1)
	if err != nil {
		t.Fatalf("List() after event error = %v", err)
	}
	updatedFirstTask := items[0]
	if updatedFirstTask.CurrentCount != 1 || updatedFirstTask.Status != models.InProgressTaskStatus {
		t.Fatalf("first task after event = %+v, want count 1 and IN_PROGRESS", updatedFirstTask)
	}
}

func TestDefinitionsForDateIsStable(t *testing.T) {
	service := testService(t)
	firstDate := time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC)
	secondDate := firstDate.AddDate(0, 0, 1)

	firstSelection := service.definitionsForDate(firstDate)
	repeatedSelection := service.definitionsForDate(firstDate)
	if len(firstSelection) != TotalDailyTasks {
		t.Fatalf("first selection has %d tasks, want %d", len(firstSelection), TotalDailyTasks)
	}
	for index := range firstSelection {
		if firstSelection[index] != repeatedSelection[index] {
			t.Fatalf("selection changed for the same date: %v and %v", firstSelection, repeatedSelection)
		}
		if firstSelection[index].Slot != index+1 {
			t.Errorf("selection %d slot = %d, want %d", index, firstSelection[index].Slot, index+1)
		}
	}

	secondSelection := service.definitionsForDate(secondDate)
	if len(secondSelection) != TotalDailyTasks {
		t.Fatalf("second selection has %d tasks, want %d", len(secondSelection), TotalDailyTasks)
	}
}

func TestProgressCountsCompletedAndClaimedTasks(t *testing.T) {
	service := testService(t)
	userID := uuid.New()

	if _, err := service.List(context.Background(), userID, 1); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	firstTask := testTaskBySlot(t, service, 1)
	secondTask := testTaskBySlot(t, service, 2)

	err := service.RecordEvents(context.Background(), userID, []Event{
		{Type: firstTask.Type, Count: firstTask.TargetCount},
		{Type: secondTask.Type, Count: secondTask.TargetCount},
	}, 1)
	if err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}
	if _, err := service.Claim(context.Background(), userID, firstTask.ID, 1); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}

	progress, err := service.Progress(context.Background(), userID, 1)
	if err != nil {
		t.Fatalf("Progress() error = %v", err)
	}
	if progress.CompletedCount != 2 || progress.TotalCount != TotalDailyTasks {
		t.Fatalf("Progress() = %+v, want 2/%d", progress, TotalDailyTasks)
	}
}

func TestRecordEventsEmptyBatchDoesNothing(t *testing.T) {
	service := testService(t)

	if err := service.RecordEvents(context.Background(), uuid.New(), nil, 1); err != nil {
		t.Fatalf("RecordEvents() error = %v", err)
	}

	var count int64
	if err := service.db.Model(&models.Task{}).Count(&count).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if count != 0 {
		t.Fatalf("tasks after empty batch = %d, want 0", count)
	}
}

func TestClaimReturnsTaskNotFound(t *testing.T) {
	service := testService(t)

	_, err := service.Claim(context.Background(), uuid.New(), uuid.New(), 1)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("Claim() error = %v, want ErrTaskNotFound", err)
	}
}

func TestDefinitionsRejectInvalidValues(t *testing.T) {
	base, err := LoadDefaultDefinitions()
	if err != nil {
		t.Fatalf("LoadDefaultDefinitions() error = %v", err)
	}

	tests := []struct {
		name string
		edit func([]Definition) []Definition
	}{
		{name: "unknown slot", edit: func(definitions []Definition) []Definition {
			definitions[0].Slot = 9
			return definitions
		}},
		{name: "unknown type", edit: func(definitions []Definition) []Definition {
			definitions[0].Type = "UNKNOWN"
			return definitions
		}},
		{name: "empty description", edit: func(definitions []Definition) []Definition {
			definitions[0].Description = " "
			return definitions
		}},
		{name: "invalid target", edit: func(definitions []Definition) []Definition {
			definitions[0].TargetCount = 0
			return definitions
		}},
		{name: "wrong reward", edit: func(definitions []Definition) []Definition {
			definitions[0].RewardLeaves = 1
			return definitions
		}},
		{name: "wrong level", edit: func(definitions []Definition) []Definition {
			definitions[0].RequiredLevel = 5
			return definitions
		}},
		{name: "duplicate", edit: func(definitions []Definition) []Definition {
			return append(definitions, definitions[0])
		}},
		{name: "missing slot", edit: func(definitions []Definition) []Definition {
			result := make([]Definition, 0, len(definitions))
			for _, definition := range definitions {
				if definition.Slot != 4 {
					result = append(result, definition)
				}
			}
			return result
		}},
		{name: "missing type", edit: func(definitions []Definition) []Definition {
			result := make([]Definition, 0, len(definitions))
			for _, definition := range definitions {
				if definition.Type != models.OrderWithDeliveryTaskType {
					result = append(result, definition)
				}
			}
			return result
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definitions := append([]Definition(nil), base...)
			if err := validateDefinitions(test.edit(definitions)); err == nil {
				t.Fatal("validateDefinitions() error = nil, want error")
			}
		})
	}
}

func TestLoadDefinitionsRejectsInvalidYAML(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "malformed", data: "tasks: ["},
		{name: "unknown field", data: "tasks: []\nother: true"},
		{name: "multiple documents", data: "tasks: []\n---\ntasks: []"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := loadDefinitions([]byte(test.data)); err == nil {
				t.Fatal("loadDefinitions() error = nil, want error")
			}
		})
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

	definitions, err := LoadDefaultDefinitions()
	if err != nil {
		t.Fatalf("load task definitions: %v", err)
	}

	service := NewService(db, definitions)
	service.now = func() time.Time {
		return time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	}

	return service
}

func testTask(t *testing.T, service *Service, taskType models.TaskType) models.Task {
	t.Helper()

	var task models.Task
	if err := service.db.Where("date = ? AND type = ?", service.today(), taskType).First(&task).Error; err != nil {
		t.Fatalf("find task %s: %v", taskType, err)
	}

	return task
}

func testTaskBySlot(t *testing.T, service *Service, slot int) models.Task {
	t.Helper()

	var task models.Task
	if err := service.db.Where("date = ? AND slot = ?", service.today(), slot).First(&task).Error; err != nil {
		t.Fatalf("find task in slot %d: %v", slot, err)
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
