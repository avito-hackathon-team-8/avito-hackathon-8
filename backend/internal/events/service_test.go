package events

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/tasks"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var eventTestNow = time.Now().UTC().Truncate(time.Second)

func TestTaskEventIsIdempotentAndConflictingReplayFails(t *testing.T) {
	service, db, userID, assignment := testEventService(t, 1)
	eventID := uuid.New()
	event := Event{ID: eventID, Type: string(models.ViewListingsTaskType), Count: 2, OccurredAt: eventTestNow.Add(-time.Minute)}
	if err := service.Record(context.Background(), userID, []Event{event}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if err := service.Record(context.Background(), userID, []Event{event}); err != nil {
		t.Fatalf("identical replay error = %v", err)
	}
	service.now = func() time.Time { return eventTestNow.AddDate(0, 0, 1) }
	if err := service.Record(context.Background(), userID, []Event{event}); err != nil {
		t.Fatalf("identical replay after UTC day changed error = %v", err)
	}

	var stored models.UserDailyTask
	if err := db.First(&stored, "id = ?", assignment.ID).Error; err != nil {
		t.Fatalf("load assignment: %v", err)
	}
	if stored.CurrentCount != 2 {
		t.Fatalf("task count after replay = %d, want 2", stored.CurrentCount)
	}

	conflict := event
	conflict.Count = 3
	if err := service.Record(context.Background(), userID, []Event{conflict}); !errors.Is(err, ErrEventConflict) {
		t.Fatalf("conflicting replay error = %v, want ErrEventConflict", err)
	}
}

func TestLoginAcceptsCurrentISOWeekAndRejectsFutureOrOldEvents(t *testing.T) {
	service, db, userID, _ := testEventService(t, 1)
	today := utcDate(eventTestNow)
	monday := today.AddDate(0, 0, -(int(today.Weekday())+6)%7).Add(8 * time.Hour)
	if err := service.Record(context.Background(), userID, []Event{{ID: uuid.New(), Type: LoginType, Count: 1, OccurredAt: monday}}); err != nil {
		t.Fatalf("late current-week LOGIN error = %v", err)
	}
	var login models.UserLogin
	if err := db.First(&login, "user_id = ?", userID).Error; err != nil {
		t.Fatalf("load login: %v", err)
	}
	if !utcDate(login.ActivityDate).Equal(utcDate(monday)) {
		t.Fatalf("login date = %v, want %v", login.ActivityDate, monday)
	}

	tests := []Event{
		{ID: uuid.New(), Type: LoginType, Count: 1, OccurredAt: eventTestNow.Add(time.Second)},
		{ID: uuid.New(), Type: LoginType, Count: 1, OccurredAt: monday.AddDate(0, 0, -1)},
		{ID: uuid.New(), Type: string(models.ViewListingsTaskType), Count: 1, OccurredAt: eventTestNow.AddDate(0, 0, -1)},
	}
	for _, event := range tests {
		if err := service.Record(context.Background(), userID, []Event{event}); !errors.Is(err, ErrEventOutsideTime) {
			t.Errorf("Record(%+v) error = %v, want ErrEventOutsideTime", event, err)
		}
	}
}

func TestBatchRollsBackAllEventsWhenTaskIsLocked(t *testing.T) {
	service, db, userID, _ := testEventService(t, 5)
	loginID := uuid.New()
	taskID := uuid.New()
	err := service.Record(context.Background(), userID, []Event{
		{ID: loginID, Type: LoginType, Count: 1, OccurredAt: eventTestNow.Add(-time.Minute)},
		{ID: taskID, Type: string(models.ViewListingsTaskType), Count: 1, OccurredAt: eventTestNow.Add(-time.Minute)},
	})
	if !errors.Is(err, tasks.ErrTaskLocked) {
		t.Fatalf("Record() error = %v, want ErrTaskLocked", err)
	}
	var eventCount, loginCount int64
	if err := db.Model(&models.ExternalEvent{}).Where("id IN ?", []uuid.UUID{loginID, taskID}).Count(&eventCount).Error; err != nil {
		t.Fatalf("count external events: %v", err)
	}
	if err := db.Model(&models.UserLogin{}).Where("user_id = ?", userID).Count(&loginCount).Error; err != nil {
		t.Fatalf("count logins: %v", err)
	}
	if eventCount != 0 || loginCount != 0 {
		t.Fatalf("rolled-back rows: events=%d logins=%d", eventCount, loginCount)
	}
}

func TestRecordRejectsInvalidBatchBeforeWriting(t *testing.T) {
	service, db, userID, _ := testEventService(t, 1)
	validID := uuid.New()
	err := service.Record(context.Background(), userID, []Event{
		{ID: validID, Type: LoginType, Count: 1, OccurredAt: eventTestNow.Add(-time.Minute)},
		{ID: uuid.New(), Type: "UNKNOWN", Count: 1, OccurredAt: eventTestNow.Add(-time.Minute)},
	})
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("Record() error = %v, want ErrInvalidEvent", err)
	}
	err = service.Record(context.Background(), userID, []Event{
		{ID: uuid.New(), Type: string(models.ViewListingsTaskType), Count: 1, OccurredAt: eventTestNow.Add(-time.Minute)},
		{ID: uuid.New(), Type: LoginType, Count: 1, OccurredAt: eventTestNow.Add(time.Second)},
	})
	if !errors.Is(err, ErrEventOutsideTime) {
		t.Fatalf("future event after task error = %v, want ErrEventOutsideTime", err)
	}
	var count int64
	if err := db.Model(&models.ExternalEvent{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("external events after invalid batch = %d, error = %v", count, err)
	}
}

func testEventService(t *testing.T, unlockLevel int) (*Service, *gorm.DB, uuid.UUID, models.UserDailyTask) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Pet{}, &models.DailyTaskDefinition{}, &models.UserDailyTask{}, &models.UserLogin{}, &models.ExternalEvent{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString()), Verified: true, CreatedAt: eventTestNow.AddDate(0, 0, -10)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.Pet{UserID: user.ID, Level: 1}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	day := utcDate(eventTestNow)
	taskTypes := []models.TaskType{models.ViewListingsTaskType, models.AddToFavoritesTaskType, models.PublishListingTaskType, models.CompleteDealTaskType}
	var assignment models.UserDailyTask
	for index, taskType := range taskTypes {
		definition := models.DailyTaskDefinition{Code: fmt.Sprintf("task-%d", index+1), Type: taskType, Title: fmt.Sprintf("Task %d", index+1), Slot: index + 1, TargetCount: 3, Reward: 10, UnlockLevel: unlockLevel, Categories: "[]", Active: true}
		if err := db.Create(&definition).Error; err != nil {
			t.Fatalf("create definition: %v", err)
		}
		status := models.InProgressTaskStatus
		if unlockLevel > 1 {
			status = models.LockedTaskStatus
		}
		created := models.UserDailyTask{UserID: user.ID, TaskDefinitionID: definition.ID, Day: day, Status: status, ExpiresAt: day.AddDate(0, 0, 1)}
		if err := db.Create(&created).Error; err != nil {
			t.Fatalf("create assignment: %v", err)
		}
		if index == 0 {
			assignment = created
		}
	}
	notifier := testutil.DailyReportNotifierMock{}
	taskService := tasks.NewService(db, notifier)
	service := NewService(db, notifier, taskService)
	service.now = func() time.Time { return eventTestNow }
	return service, db, user.ID, assignment
}
