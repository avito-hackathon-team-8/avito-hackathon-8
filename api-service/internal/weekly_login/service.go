package weekly_login

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/metrics"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/petstate"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUserNotFound         = errors.New("player not found")
	ErrAlreadyClaimed       = errors.New("weekly login reward already claimed")
	ErrActivityNotConfirmed = errors.New("activity is not confirmed")
)

const (
	weeklyLoginDays     = 7
	petStateReadTimeout = 750 * time.Millisecond
)

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
	db          *gorm.DB
	now         func() time.Time
	pets        *pet.Service
	dailyReport DailyReportNotifier
	state       HappinessProvider
	metrics     *appmetrics.Metrics
}

func (service *Service) SetMetrics(metrics *appmetrics.Metrics) { service.metrics = metrics }

type HappinessProvider interface {
	Get(context.Context, uuid.UUID) (petstate.Snapshot, error)
}

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
}

func NewService(db *gorm.DB, dailyReport DailyReportNotifier, petService *pet.Service, state ...HappinessProvider) *Service {
	service := &Service{db: db, now: time.Now, pets: petService, dailyReport: dailyReport}

	if len(state) > 0 {
		service.state = state[0]
	}

	return service
}

type ClaimResult struct {
	Claim    models.WeeklyLoginClaim
	Progress pet.Progress
}

type WeeklyLoginDay struct {
	Weekday             int
	Date                string
	Status              models.DayStatus
	RewardLeaves        int
	BaseRewardLeaves    int
	HappinessMultiplier float64
	ClaimID             *uuid.UUID
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
		activityInactive = service.checkLoginDate(service.db.WithContext(ctx), userID, today)
	}

	week := buildCurrentWeek(user, claims, today, activityInactive)

	if service.state == nil {
		return week, nil
	}

	stateCtx, cancel := context.WithTimeout(ctx, petStateReadTimeout)
	snapshot, err := service.state.Get(stateCtx, userID)
	cancel()

	if err != nil {
		if service.metrics != nil {
			service.metrics.Fallbacks.WithLabelValues("pet-state", "weekly-login-get").Inc()
		}

		log.Printf("service=api-service operation=weekly-login-get error_type=pet-state-fallback: %v", err)

		return week, nil
	}

	for index := range week.Claims {
		day := &week.Claims[index]

		if day.Status == models.DayStatusClaimed || day.RewardLeaves == 0 {
			continue
		}

		day.BaseRewardLeaves = day.RewardLeaves
		day.HappinessMultiplier = snapshot.HappinessMultiplier
		day.RewardLeaves = calculateReward(day.BaseRewardLeaves, snapshot.Happiness)
	}

	return week, nil
}

func (service *Service) Claim(ctx context.Context, userID uuid.UUID) (ClaimResult, error) {
	claimDate := utcDate(service.now())

	var happiness *float64
	var multiplier *float64

	if service.state != nil {
		stateCtx, cancel := context.WithTimeout(ctx, petStateReadTimeout)
		snapshot, err := service.state.Get(stateCtx, userID)
		cancel()

		if err != nil {
			if service.metrics != nil {
				service.metrics.Fallbacks.WithLabelValues("pet-state", "weekly-login-claim").Inc()
			}

			log.Printf("weekly login pet state fallback user_id=%s operation=claim: %v", userID, err)
		} else {
			happiness = &snapshot.Happiness
			multiplier = &snapshot.HappinessMultiplier
		}
	}

	var claim models.WeeklyLoginClaim
	var progress pet.Progress

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

		if service.checkLoginDate(tx, userID, claimDate) {
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

		baseReward := int(reward)
		actualReward := baseReward

		if happiness != nil {
			actualReward = calculateReward(baseReward, *happiness)
		}

		claim = models.WeeklyLoginClaim{
			UserID:              userID,
			ClaimDate:           claimDate,
			RewardLeaves:        actualReward,
			BaseRewardLeaves:    baseReward,
			HappinessSnapshot:   happiness,
			HappinessMultiplier: multiplier,
		}

		if err := tx.Create(&claim).Error; err != nil {
			return err
		}

		if service.pets == nil {
			return errors.New("pet service is not configured")
		}

		progress, err = service.pets.CreditTx(tx, pet.Credit{
			UserID: userID, Amount: int64(actualReward), Reason: models.LeafReasonWeeklyLogin,
			OperationKey: fmt.Sprintf("weekly-login:%s", claim.ID),
		})

		return err
	})

	if errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrAlreadyClaimed) ||
		errors.Is(err, ErrActivityNotConfirmed) {
		return ClaimResult{}, err
	}

	if err != nil {
		return ClaimResult{}, fmt.Errorf("claim weekly login reward: %w", err)
	}

	service.dailyReport.Notify(userID)

	return ClaimResult{Claim: claim, Progress: progress}, nil
}

func (service *Service) RecordToday(ctx context.Context, userID uuid.UUID) error {
	now := service.now().UTC()
	activity := models.UserLogin{
		UserID:       userID,
		ActivityDate: utcDate(now),
		CreatedAt:    now,
	}

	err := service.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "activity_date"},
			},
			DoNothing: true,
		}).
		Create(&activity).Error

	if err != nil {
		return fmt.Errorf("record today's user activity: %w", err)
	}

	service.dailyReport.Notify(userID)

	return nil
}

func (service *Service) checkLoginDate(db *gorm.DB, userID uuid.UUID, date time.Time) bool {
	var count int64

	err := db.Model(&models.UserLogin{}).
		Where("user_id = ? AND activity_date = ?", userID, utcDate(date)).
		Count(&count).Error

	return err != nil || count == 0
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

func buildCurrentWeek(user models.User, claims []models.WeeklyLoginClaim, today time.Time, activityInactive bool) CurrentWeek {
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
			Weekday:             dayIndex + 1,
			Date:                dateString,
			HappinessMultiplier: 1,
		}

		if claim, ok := claimsByDate[dateString]; ok {
			claimID := claim.ID
			day.Status = models.DayStatusClaimed
			day.RewardLeaves = claim.RewardLeaves
			day.BaseRewardLeaves = claim.BaseRewardLeaves

			if claim.HappinessMultiplier != nil {
				day.HappinessMultiplier = *claim.HappinessMultiplier
			}

			day.ClaimID = &claimID

			result.Claims = append(result.Claims, day)

			continue
		}

		switch {
		case date.Before(registrationDate):
			day.Status = models.DayStatusMissed
		case date.Before(today):
			day.Status = models.DayStatusMissed
		case date.After(today):
			day.Status = models.DayStatusFuture
		case activityInactive:
			day.Status = models.DayStatusFuture
		default:
			day.Status = models.DayStatusAvailable
		}

		if day.Status == models.DayStatusAvailable || day.Status == models.DayStatusFuture {
			reward := weeklyRewardByIndex(nextRewardIndex)

			if reward != 0 {
				day.RewardLeaves = int(reward)
				day.BaseRewardLeaves = int(reward)
				day.HappinessMultiplier = 1

				nextRewardIndex++
			}
		}

		result.Claims = append(result.Claims, day)
	}

	return result
}

func calculateReward(base int, happiness float64) int {
	multiplier := 0.5 + math.Max(0, math.Min(100, happiness))/100

	return int(math.Round(float64(base) * multiplier))
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
