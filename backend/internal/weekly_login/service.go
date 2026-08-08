package weekly_login

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUserNotFound         = errors.New("player not found")
	ErrAlreadyClaimed       = errors.New("weekly login reward already claimed")
	ErrActivityNotConfirmed = errors.New("activity is not confirmed")
)

const weeklyLoginDays = 7

// Награда за каждый день недели
type weeklyReward int

const (
	weeklyRewardFirst   weeklyReward = 10
	weeklyRewardSecond  weeklyReward = 20
	weeklyRewardThird   weeklyReward = 30
	weeklyRewardFourth  weeklyReward = 40
	weeklyRewardFifth   weeklyReward = 50
	weeklyRewardSixth   weeklyReward = 60
	weeklyRewardSeventh weeklyReward = 70
)

type Service struct {
	db                  *gorm.DB
	now                 func() time.Time
	activity            ActivityProvider
	dailyReportNotifier DailyReportNotifier
}

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
}

func NewService(db *gorm.DB, activity ActivityProvider, dailyReportNotifier DailyReportNotifier) *Service {
	return &Service{
		db:                  db,
		now:                 time.Now,
		activity:            activity,
		dailyReportNotifier: dailyReportNotifier,
	}
}

type WeeklyLoginDay struct {
	Weekday      int
	Date         string
	Status       models.DayStatus
	RewardLeaves int
	ClaimID      *uuid.UUID
}

type CurrentWeek struct {
	ClaimedDaysCount int
	Claims           []WeeklyLoginDay
}

func (service *Service) Get(ctx context.Context, userID uuid.UUID) (CurrentWeek, error) {
	var user models.User

	err := service.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CurrentWeek{}, ErrUserNotFound
	}

	if err != nil {
		return CurrentWeek{}, fmt.Errorf("get weekly login user: %w", err)
	}

	today := utcDate(service.now())
	weekStart, weekEnd := utcWeekBounds(today)
	var claims []models.WeeklyLoginClaim

	err = service.db.WithContext(ctx).
		Where("user_id = ? AND claim_date >= ? AND claim_date < ?", userID, weekStart, weekEnd).
		Find(&claims).Error

	if err != nil {
		return CurrentWeek{}, fmt.Errorf("get weekly login claims: %w", err)
	}

	activityInactive := false

	if shouldCheckTodayActivity(user, claims, today) {
		activityInactive = service.activityInactive(ctx, userID, today)
	}

	return buildCurrentWeek(user, claims, today, activityInactive), nil
}

func (service *Service) Claim(ctx context.Context, userID uuid.UUID, date time.Time) (models.WeeklyLoginClaim, error) {
	claimDate := utcDate(date)
	var claim models.WeeklyLoginClaim

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).
			First(&user).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}

		if err != nil {
			return err
		}

		err = tx.Where("user_id = ? AND claim_date = ?", userID, claimDate).
			First(&claim).Error

		if err == nil {
			return ErrAlreadyClaimed
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if service.activityInactive(ctx, userID, claimDate) {
			return ErrActivityNotConfirmed
		}

		weekStart, weekEnd := utcWeekBounds(claimDate)
		var claimedDaysCount int64

		if err := tx.Model(&models.WeeklyLoginClaim{}).
			Where("user_id = ? AND claim_date >= ? AND claim_date < ?", userID, weekStart, weekEnd).
			Count(&claimedDaysCount).Error; err != nil {
			return err
		}

		reward := weeklyRewardByIndex(int(claimedDaysCount))

		if reward == 0 {
			return fmt.Errorf("weekly reward is not configured for claim index %d", claimedDaysCount)
		}

		claim = models.WeeklyLoginClaim{
			UserID:       userID,
			ClaimDate:    claimDate,
			RewardLeaves: int(reward),
		}

		return tx.Create(&claim).Error
	})

	if errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrAlreadyClaimed) ||
		errors.Is(err, ErrActivityNotConfirmed) {
		return models.WeeklyLoginClaim{}, err
	}

	if err != nil {
		return models.WeeklyLoginClaim{}, fmt.Errorf("claim weekly login reward: %w", err)
	}

	service.dailyReportNotifier.Notify(userID)

	return claim, nil
}

func (service *Service) activityInactive(ctx context.Context, userID uuid.UUID, date time.Time) bool {
	if service.activity == nil {
		return false
	}

	activity, err := service.activity.Get(ctx, userID, date)

	if err != nil {
		return false
	}

	return !activity.Active
}

func weeklyRewardByIndex(index int) weeklyReward {
	switch index {
	case 0:
		return weeklyRewardFirst
	case 1:
		return weeklyRewardSecond
	case 2:
		return weeklyRewardThird
	case 3:
		return weeklyRewardFourth
	case 4:
		return weeklyRewardFifth
	case 5:
		return weeklyRewardSixth
	case 6:
		return weeklyRewardSeventh
	default:
		return 0
	}
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()

	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func utcWeekBounds(date time.Time) (time.Time, time.Time) {
	date = utcDate(date)
	daysSinceMonday := (int(date.Weekday()) + 6) % 7
	weekStart := date.AddDate(0, 0, -daysSinceMonday)

	return weekStart, weekStart.AddDate(0, 0, 7)
}

func buildCurrentWeek(
	user models.User,
	claims []models.WeeklyLoginClaim,
	today time.Time,
	activityInactive bool,
) CurrentWeek {
	today = utcDate(today)
	weekStart, _ := utcWeekBounds(today)
	registrationDate := utcDate(user.CreatedAt)
	claimsByDate := make(map[string]models.WeeklyLoginClaim, len(claims))

	for _, claim := range claims {
		claimsByDate[utcDate(claim.ClaimDate).Format(time.DateOnly)] = claim
	}

	result := CurrentWeek{
		ClaimedDaysCount: len(claims),
		Claims:           make([]WeeklyLoginDay, 0, weeklyLoginDays),
	}
	nextRewardIndex := len(claims)

	for dayIndex := range weeklyLoginDays {
		date := weekStart.AddDate(0, 0, dayIndex)
		dateString := date.Format(time.DateOnly)
		day := WeeklyLoginDay{
			Weekday: dayIndex + 1,
			Date:    dateString,
		}

		if claim, ok := claimsByDate[dateString]; ok {
			claimID := claim.ID
			day.Status = models.DayStatusClaimed
			day.RewardLeaves = claim.RewardLeaves
			day.ClaimID = &claimID
			result.Claims = append(result.Claims, day)

			continue
		}

		switch {
		case date.Before(registrationDate):
			day.Status = models.DayStatusBeforeRegistration
		case date.Before(today):
			day.Status = models.DayStatusMissed
		case date.After(today):
			day.Status = models.DayStatusFuture
		case activityInactive:
			day.Status = models.DayStatusUnconfirmed
		default:
			day.Status = models.DayStatusAvailable
		}

		if day.Status == models.DayStatusAvailable || day.Status == models.DayStatusFuture {
			reward := weeklyRewardByIndex(nextRewardIndex)

			if reward != 0 {
				day.RewardLeaves = int(reward)
				nextRewardIndex++
			}
		}

		result.Claims = append(result.Claims, day)
	}

	return result
}

func shouldCheckTodayActivity(user models.User, claims []models.WeeklyLoginClaim, today time.Time) bool {
	if today.Before(utcDate(user.CreatedAt)) {
		return false
	}

	for _, claim := range claims {
		if utcDate(claim.ClaimDate).Equal(today) {
			return false
		}
	}

	return true
}
