package daily_report

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGetReturnsEmptyReportAndRejectsUnknownUser(t *testing.T) {
	now := time.Date(2026, time.August, 5, 2, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	service, _, user := testDailyReportService(t, now)

	report, err := service.Get(context.Background(), user.ID)

	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	wantDayStart := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)

	if report.LeavesEarnedToday != 0 || report.Date != "2026-08-04" || report.LevelUp != nil || report.VisitedToday {
		t.Fatalf("empty report = %+v", report)
	}

	if report.Rewards == nil || len(report.Rewards) != 0 {
		t.Fatalf("Rewards = %+v, want non-nil empty slice", report.Rewards)
	}

	if report.Tasks == nil || len(report.Tasks) != 0 {
		t.Fatalf("Tasks = %+v, want non-nil empty slice", report.Tasks)
	}

	if !report.UpdatedAt.Equal(wantDayStart) {
		t.Fatalf("UpdatedAt = %s, want %s", report.UpdatedAt, wantDayStart)
	}

	tests := []struct {
		name   string
		userID uuid.UUID
	}{
		{name: "nil user", userID: uuid.Nil},
		{name: "missing user", userID: uuid.New()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.Get(context.Background(), test.userID); !errors.Is(err, ErrUserNotFound) {
				t.Fatalf("Get() error = %v, want ErrUserNotFound", err)
			}
		})
	}
}

func TestGetBuildsDailyReportFromTodaysActivity(t *testing.T) {
	now := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	service, db, user := testDailyReportService(t, now)
	dayStart := time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	firstReward := models.Reward{
		UserID:    user.ID,
		Title:     "Скидка на доставку",
		Category:  models.RewardCategoryDeliveryDiscount,
		Source:    models.RewardSourceLeaderboard,
		ExpiresAt: dayStart.AddDate(0, 1, 0),
		CreatedAt: dayStart.Add(8 * time.Hour),
	}

	secondReward := models.Reward{
		UserID:    user.ID,
		Title:     "Бонусы Авито",
		Category:  models.RewardCategoryAvitoBonus,
		Source:    models.RewardSourceLeaderboard,
		ExpiresAt: dayStart.AddDate(0, 2, 0),
		CreatedAt: dayStart.Add(10 * time.Hour),
	}

	if err := db.Create(&secondReward).Error; err != nil {
		t.Fatalf("create second reward: %v", err)
	}

	if err := db.Create(&firstReward).Error; err != nil {
		t.Fatalf("create first reward: %v", err)
	}

	completedAt := dayStart.Add(9 * time.Hour)
	claimedAt := dayStart.Add(11*time.Hour + 30*time.Minute)
	secondTask := seedDailyReportTask(t, db, user.ID, 2, models.AddToFavoritesTaskType, "Добавить в избранное", 20, completedAt, nil)
	firstTask := seedDailyReportTask(t, db, user.ID, 1, models.ViewListingsTaskType, "Посмотреть объявления", 45, completedAt, &claimedAt)

	login := models.UserLogin{
		UserID:       user.ID,
		ActivityDate: dayStart,
		CreatedAt:    dayStart.Add(7 * time.Hour),
	}

	if err := db.Create(&login).Error; err != nil {
		t.Fatalf("create user activity: %v", err)
	}

	transactions := []models.LeafTransaction{
		{UserID: user.ID, Amount: 45, Reason: models.LeafReasonTaskReward, OperationKey: "task-reward", OccurredAt: dayStart.Add(9*time.Hour + 15*time.Minute)},
		{UserID: user.ID, Amount: 10, Reason: models.LeafReasonWeeklyLogin, OperationKey: "weekly-login", OccurredAt: dayStart.Add(9*time.Hour + 30*time.Minute)},
		{UserID: user.ID, Amount: 0, Reason: models.LeafReasonLevelUp, OperationKey: "level-up-1", OccurredAt: dayStart.Add(10*time.Hour + 15*time.Minute)},
		{UserID: user.ID, Amount: 0, Reason: models.LeafReasonLevelUp, OperationKey: "level-up-2", OccurredAt: dayStart.Add(10*time.Hour + 45*time.Minute)},
		{UserID: user.ID, Amount: -models.ChestOpeningLeavesCost, Reason: models.LeafReasonChestPurchase, OperationKey: "chest-purchase", OccurredAt: dayStart.Add(11*time.Hour + 15*time.Minute)},
	}

	if err := db.Create(&transactions).Error; err != nil {
		t.Fatalf("create leaf transactions: %v", err)
	}

	if err := db.Create(&models.Pet{UserID: user.ID, Level: 4}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	report, err := service.Get(context.Background(), user.ID)

	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if report.LeavesEarnedToday != 55 || report.Date != "2026-08-05" || !report.VisitedToday {
		t.Fatalf("report summary = %+v", report)
	}

	if len(report.Rewards) != 2 {
		t.Fatalf("len(Rewards) = %d, want 2", len(report.Rewards))
	}

	if report.Rewards[0].ID != firstReward.ID || report.Rewards[1].ID != secondReward.ID {
		t.Fatalf("Rewards order = %+v, want first reward followed by second reward", report.Rewards)
	}

	if report.Rewards[0].Type != firstReward.Category || report.Rewards[0].Title != firstReward.Title ||
		!report.Rewards[0].ExpiresAt.Equal(firstReward.ExpiresAt) || !report.Rewards[0].ReceivedAt.Equal(firstReward.CreatedAt) {
		t.Fatalf("first reward = %+v", report.Rewards[0])
	}

	if len(report.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, want only the claimed task", len(report.Tasks))
	}

	if report.Tasks[0].ID != firstTask.ID {
		t.Fatalf("Tasks = %+v, want only claimed task %s; unclaimed task %s must be excluded", report.Tasks, firstTask.ID, secondTask.ID)
	}

	if report.Tasks[0].Type != models.ViewListingsTaskType || report.Tasks[0].Description != "Посмотреть объявления" ||
		report.Tasks[0].RewardLeaves != 45 || !report.Tasks[0].CompletedAt.Equal(completedAt) {

		t.Fatalf("first task = %+v", report.Tasks[0])
	}

	if report.LevelUp == nil || report.LevelUp.FromLevel != 2 || report.LevelUp.ToLevel != 4 ||
		!report.LevelUp.OccurredAt.Equal(transactions[3].OccurredAt) {
		t.Fatalf("LevelUp = %+v, want levels 2 to 4 at last level-up time %s", report.LevelUp, transactions[3].OccurredAt)
	}

	if !report.UpdatedAt.Equal(claimedAt) {
		t.Fatalf("UpdatedAt = %s, want %s", report.UpdatedAt, claimedAt)
	}

	if !report.UpdatedAt.Before(dayEnd) {
		t.Fatalf("UpdatedAt = %s, want current UTC day", report.UpdatedAt)
	}
}

func TestGetUsesUTCDateBoundsAndIsolatesUsers(t *testing.T) {
	now := time.Date(2026, time.August, 5, 2, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	service, db, user := testDailyReportService(t, now)
	dayStart := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	otherUser := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}

	if err := db.Create(&otherUser).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}

	includedReward := models.Reward{UserID: user.ID, Title: "Included", Category: models.RewardCategoryAvitoBonus, Source: models.RewardSourceLeaderboard, ExpiresAt: dayEnd.AddDate(0, 1, 0), CreatedAt: dayStart}
	excludedRewards := []models.Reward{
		{UserID: user.ID, Title: "Previous day", Category: models.RewardCategoryAvitoBonus, Source: models.RewardSourceLeaderboard, ExpiresAt: dayEnd, CreatedAt: dayStart.Add(-time.Second)},
		{UserID: user.ID, Title: "Next day", Category: models.RewardCategoryAvitoBonus, Source: models.RewardSourceLeaderboard, ExpiresAt: dayEnd.AddDate(0, 1, 0), CreatedAt: dayEnd},
		{UserID: otherUser.ID, Title: "Other user", Category: models.RewardCategoryAvitoBonus, Source: models.RewardSourceLeaderboard, ExpiresAt: dayEnd, CreatedAt: dayStart.Add(time.Hour)},
	}

	if err := db.Create(&includedReward).Error; err != nil {
		t.Fatalf("create included reward: %v", err)
	}

	if err := db.Create(&excludedRewards).Error; err != nil {
		t.Fatalf("create excluded rewards: %v", err)
	}

	includedCompletedAt := dayStart.Add(22 * time.Hour)
	includedTask := seedDailyReportTask(t, db, user.ID, 1, models.ViewListingsTaskType, "Included", 10, includedCompletedAt, &includedCompletedAt)

	seedDailyReportTask(t, db, user.ID, 2, models.AddToFavoritesTaskType, "Previous day", 10, dayStart.Add(-time.Second), nil)
	seedDailyReportTask(t, db, user.ID, 3, models.PublishListingTaskType, "Next day", 10, dayEnd, nil)
	seedDailyReportTask(t, db, otherUser.ID, 1, models.ViewListingsTaskType, "Other user", 10, dayStart.Add(21*time.Hour), nil)

	transactions := []models.LeafTransaction{
		{UserID: user.ID, Amount: 5, Reason: models.LeafReasonTaskReward, OperationKey: "included", OccurredAt: dayStart},
		{UserID: user.ID, Amount: 100, Reason: models.LeafReasonTaskReward, OperationKey: "previous-day", OccurredAt: dayStart.Add(-time.Second)},
		{UserID: user.ID, Amount: 100, Reason: models.LeafReasonTaskReward, OperationKey: "next-day", OccurredAt: dayEnd},
		{UserID: otherUser.ID, Amount: 100, Reason: models.LeafReasonTaskReward, OperationKey: "other-user", OccurredAt: dayStart.Add(23 * time.Hour)},
	}

	if err := db.Create(&transactions).Error; err != nil {
		t.Fatalf("create leaf transactions: %v", err)
	}

	activities := []models.UserLogin{
		{UserID: user.ID, ActivityDate: dayStart, CreatedAt: dayStart.Add(3 * time.Hour)},
		{UserID: user.ID, ActivityDate: dayStart.AddDate(0, 0, -1), CreatedAt: dayStart.Add(-time.Hour)},
		{UserID: otherUser.ID, ActivityDate: dayStart, CreatedAt: dayStart.Add(23 * time.Hour)},
	}

	if err := db.Create(&activities).Error; err != nil {
		t.Fatalf("create user activities: %v", err)
	}

	report, err := service.Get(context.Background(), user.ID)

	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if report.Date != "2026-08-04" {
		t.Fatalf("Date = %q, want 2026-08-04", report.Date)
	}

	if len(report.Rewards) != 1 || report.Rewards[0].ID != includedReward.ID {
		t.Fatalf("Rewards = %+v, want only reward at UTC day start", report.Rewards)
	}

	if len(report.Tasks) != 1 || report.Tasks[0].ID != includedTask.ID {
		t.Fatalf("Tasks = %+v, want only current user's task inside UTC day", report.Tasks)
	}

	if report.LeavesEarnedToday != 5 {
		t.Fatalf("LeavesEarnedToday = %d, want 5", report.LeavesEarnedToday)
	}

	if !report.VisitedToday {
		t.Fatal("VisitedToday = false, want true")
	}

	if report.LevelUp != nil {
		t.Fatalf("LevelUp = %+v, want nil", report.LevelUp)
	}

	if !report.UpdatedAt.Equal(includedCompletedAt) {
		t.Fatalf("UpdatedAt = %s, want %s", report.UpdatedAt, includedCompletedAt)
	}
}

func testDailyReportService(t *testing.T, now time.Time) (*Service, *gorm.DB, models.User) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.LevelReward{}, &models.ChestOpening{}, &models.Reward{},
		&models.DailyTaskDefinition{}, &models.UserDailyTask{}, &models.UserLogin{}, &models.LeafTransaction{}, &models.LeaderboardTotal{}, &models.Pet{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewService(db)
	service.now = func() time.Time { return now }

	return service, db, user
}

func seedDailyReportTask(t *testing.T, db *gorm.DB, userID uuid.UUID, slot int, taskType models.TaskType,
	title string, reward int, completedAt time.Time, claimedAt *time.Time,
) models.UserDailyTask {
	t.Helper()

	definition := models.DailyTaskDefinition{
		Code:        uuid.NewString(),
		Title:       title,
		Slot:        slot,
		Type:        taskType,
		TargetCount: 1,
		Reward:      reward,
		UnlockLevel: 1,
		Categories:  "[]",
		Active:      true,
	}

	if err := db.Create(&definition).Error; err != nil {
		t.Fatalf("create task definition: %v", err)
	}

	status := models.CompletedTaskStatus

	if claimedAt != nil {
		status = models.ClaimedTaskStatus
	}

	day := time.Date(completedAt.UTC().Year(), completedAt.UTC().Month(), completedAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
	task := models.UserDailyTask{
		UserID:           userID,
		TaskDefinitionID: definition.ID,
		Day:              day,
		Status:           status,
		CurrentCount:     1,
		CompletedAt:      &completedAt,
		ClaimedAt:        claimedAt,
		ExpiresAt:        day.AddDate(0, 0, 1),
	}

	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create user daily task: %v", err)
	}

	return task
}
