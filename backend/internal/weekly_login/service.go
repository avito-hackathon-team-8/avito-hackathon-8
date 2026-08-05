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
	ErrPlayerNotFound       = errors.New("player not found")
	ErrAlreadyClaimed       = errors.New("weekly login reward already claimed")
	ErrActivityNotConfirmed = errors.New("activity is not confirmed")
)

// DayStatus Типы статусов дней
type DayStatus string

const (
	DayStatusClaimed            DayStatus = "CLAIMED"             // награда получена
	DayStatusAvailable          DayStatus = "AVAILABLE"           // награда доступна для получения
	DayStatusUnconfirmed        DayStatus = "UNCONFIRMED"         // активность за сегодня не подтверждена
	DayStatusMissed             DayStatus = "MISSED"              // награда пропущена за данный день
	DayStatusFuture             DayStatus = "FUTURE"              // день не наступил
	DayStatusBeforeRegistration DayStatus = "BEFORE_REGISTRATION" // пользователь зарегистрировался в середине недели и не может забрать награду
)

type ActivityStatus string

const (
	ActivityStatusActive             ActivityStatus = "ACTIVE"
	ActivityStatusInactive           ActivityStatus = "INACTIVE"
	ActivityStatusUnknown            ActivityStatus = "UNKNOWN"
	ActivityStatusFuture             ActivityStatus = "FUTURE"
	ActivityStatusBeforeRegistration ActivityStatus = "BEFORE_REGISTRATION"
)

type ActivityChecker interface {
	Status(ctx context.Context, userID uuid.UUID, date time.Time) (ActivityStatus, error)
}

type Service struct {
	db              *gorm.DB
	now             func() time.Time
	activityChecker ActivityChecker
}

func NewService(db *gorm.DB, activityCheckers ...ActivityChecker) *Service {
	service := &Service{db: db, now: time.Now}

	if len(activityCheckers) > 0 {
		service.activityChecker = activityCheckers[0]
	}

	return service
}

type WeeklyLoginDay struct {
	Weekday      int
	Date         string
	Status       DayStatus
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
		return CurrentWeek{}, ErrPlayerNotFound
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

	activityStatus := ActivityStatusUnknown

	if service.activityChecker != nil && shouldCheckTodayActivity(user, claims, today) {
		status, activityErr := service.activityChecker.Status(ctx, userID, today)

		if activityErr == nil {
			activityStatus = status
		}
	}

	return buildCurrentWeek(user, claims, today, activityStatus), nil
}

func (service *Service) Claim(
	ctx context.Context,
	userID uuid.UUID,
	date time.Time,
) (models.WeeklyLoginClaim, error) {
	claimDate := utcDate(date)
	var claim models.WeeklyLoginClaim

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).
			First(&user).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPlayerNotFound
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

		if service.activityChecker != nil {
			activityStatus, activityErr := service.activityChecker.Status(ctx, userID, claimDate)

			if activityErr == nil && activityStatus == ActivityStatusInactive {
				return ErrActivityNotConfirmed
			}
		}

		weekStart, weekEnd := utcWeekBounds(claimDate)
		var claimedDaysCount int64

		if err := tx.Model(&models.WeeklyLoginClaim{}).
			Where("user_id = ? AND claim_date >= ? AND claim_date < ?", userID, weekStart, weekEnd).
			Count(&claimedDaysCount).Error; err != nil {
			return err
		}

		claim = models.WeeklyLoginClaim{
			UserID:       userID,
			ClaimDate:    claimDate,
			RewardLeaves: weeklyRewardLadder[claimedDaysCount],
		}

		return tx.Create(&claim).Error
	})

	if errors.Is(err, ErrPlayerNotFound) ||
		errors.Is(err, ErrAlreadyClaimed) ||
		errors.Is(err, ErrActivityNotConfirmed) {
		return models.WeeklyLoginClaim{}, err
	}

	if err != nil {
		return models.WeeklyLoginClaim{}, fmt.Errorf("claim weekly login reward: %w", err)
	}

	return claim, nil
}

var weeklyRewardLadder = [...]int{
	10,
	20,
	30,
	40,
	50,
	60,
	70,
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
	activityStatus ActivityStatus,
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
		Claims:           make([]WeeklyLoginDay, 0, len(weeklyRewardLadder)),
	}
	nextRewardIndex := len(claims)

	for dayIndex := range weeklyRewardLadder {
		date := weekStart.AddDate(0, 0, dayIndex)
		dateString := date.Format(time.DateOnly)
		day := WeeklyLoginDay{
			Weekday: dayIndex + 1,
			Date:    dateString,
		}

		if claim, ok := claimsByDate[dateString]; ok {
			claimID := claim.ID
			day.Status = DayStatusClaimed
			day.RewardLeaves = claim.RewardLeaves
			day.ClaimID = &claimID
			result.Claims = append(result.Claims, day)

			continue
		}

		switch {
		case date.Before(registrationDate):
			day.Status = DayStatusBeforeRegistration
		case date.Before(today):
			day.Status = DayStatusMissed
		case date.After(today):
			day.Status = DayStatusFuture
		case activityStatus == ActivityStatusInactive:
			day.Status = DayStatusUnconfirmed
		default:
			day.Status = DayStatusAvailable
		}

		if day.Status == DayStatusAvailable || day.Status == DayStatusFuture {
			day.RewardLeaves = weeklyRewardLadder[nextRewardIndex]
			nextRewardIndex++
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
