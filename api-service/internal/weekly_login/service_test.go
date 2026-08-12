package weekly_login

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func TestBuildCurrentWeekMarksDaysBeforeRegistrationMissed(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	user := models.User{CreatedAt: today}

	week := buildCurrentWeek(user, nil, today, false)

	if week.Claims[0].Status != models.DayStatusMissed ||
		week.Claims[1].Status != models.DayStatusMissed {

		t.Fatalf("days before registration = (%q, %q), want MISSED", week.Claims[0].Status, week.Claims[1].Status)
	}

	if week.Claims[2].Status != models.DayStatusAvailable || week.Claims[2].RewardLeaves != 10 {
		t.Fatalf("registration day = (%q, %d), want (AVAILABLE, 10)", week.Claims[2].Status, week.Claims[2].RewardLeaves)
	}
}

func TestBuildCurrentWeekMarksInactiveTodayFuture(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	user := models.User{CreatedAt: time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)}

	week := buildCurrentWeek(user, nil, today, true)

	if week.Claims[2].Status != models.DayStatusFuture || week.Claims[2].RewardLeaves != 10 {
		t.Fatalf("inactive today = (%q, %d), want (FUTURE, 10)", week.Claims[2].Status, week.Claims[2].RewardLeaves)
	}

	if week.Claims[3].Status != models.DayStatusFuture || week.Claims[3].RewardLeaves != 20 {
		t.Fatalf("next day = (%q, %d), want (FUTURE, 20)", week.Claims[3].Status, week.Claims[3].RewardLeaves)
	}
}

func TestRecordTodayAndReadActivityFromDatabase(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.UserLogin{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	now := time.Date(2026, time.August, 5, 2, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString()), CreatedAt: now}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	notifier := testutil.DailyReportNotifierMock{}
	service := NewService(db, notifier, nil)
	service.now = func() time.Time { return now }
	wantDate := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)

	if !service.checkLoginDate(db, user.ID, wantDate) {
		t.Fatal("checkLoginDate() = false before activity was recorded, want true")
	}

	if err := service.RecordToday(t.Context(), user.ID); err != nil {
		t.Fatalf("RecordToday() error = %v", err)
	}

	if err := service.RecordToday(t.Context(), user.ID); err != nil {
		t.Fatalf("second RecordToday() error = %v", err)
	}

	if service.checkLoginDate(db, user.ID, wantDate) {
		t.Fatal("checkLoginDate() = true after activity was recorded, want false")
	}

	var activities []models.UserLogin

	if err := db.Where("user_id = ?", user.ID).Find(&activities).Error; err != nil {
		t.Fatalf("load activities: %v", err)
	}

	if len(activities) != 1 || !utcDate(activities[0].ActivityDate).Equal(wantDate) {
		t.Fatalf("activities = %+v, want one activity for %s", activities, wantDate)
	}
}

func TestClaimCreditsLeavesAndLedgerAtomically(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Pet{}, &models.UserLogin{}, &models.WeeklyLoginClaim{}, &models.LeafTransaction{}, &models.LeaderboardTotal{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString()), Verified: true, CreatedAt: now.AddDate(0, 0, -1)}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.Create(&models.UserLogin{UserID: user.ID, ActivityDate: utcDate(now), CreatedAt: now}).Error; err != nil {
		t.Fatalf("create user activity: %v", err)
	}

	notifier := testutil.DailyReportNotifierMock{}
	petService := pet.NewService(db, notifier)
	service := NewService(db, notifier, petService, failingHappinessProvider{})
	service.now = func() time.Time { return now }

	result, err := service.Claim(context.Background(), user.ID)

	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}

	if !result.Claim.ClaimDate.Equal(utcDate(now)) {
		t.Fatalf("ClaimDate = %s, want current UTC date %s", result.Claim.ClaimDate, utcDate(now))
	}

	if result.Claim.RewardLeaves != 10 || result.Progress.Level != pet.InitialPetLevel || result.Progress.Leaves != pet.InitialPetLeaves+10 || result.Progress.LevelUp {
		t.Fatalf("claim result = %+v", result)
	}

	if result.Claim.HappinessSnapshot != nil || result.Claim.HappinessMultiplier != nil {
		t.Fatalf("fallback claim stored happiness data: %+v", result.Claim)
	}

	var transactions []models.LeafTransaction

	if err := db.Order("created_at").Find(&transactions).Error; err != nil {
		t.Fatalf("load ledger: %v", err)
	}

	if len(transactions) != 1 || transactions[0].Reason != models.LeafReasonWeeklyLogin || transactions[0].Amount != 10 {
		t.Fatalf("ledger = %+v", transactions)
	}

	if _, err := service.Claim(context.Background(), user.ID); !errors.Is(err, ErrAlreadyClaimed) {
		t.Fatalf("second Claim() error = %v, want ErrAlreadyClaimed", err)
	}

	var count int64

	if err := db.Model(&models.LeafTransaction{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("ledger count after replay = %d, error = %v", count, err)
	}
}
