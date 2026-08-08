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

// Максимальная длина диапазона дней
const maxActivityRangeDays = 7

var ErrInvalidActivityRange = errors.New("activity range must contain between 1 and 7 UTC dates")

type ActivityProvider interface {
	Add(ctx context.Context, userID uuid.UUID, days []ActivityDay) error
	Get(ctx context.Context, userID uuid.UUID, date time.Time) (ActivityDay, error)
	GetRange(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]ActivityDay, error)
}

type ActivityDay struct {
	Date   time.Time
	Active bool
}

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
}

type ActivityService struct {
	db          *gorm.DB
	now         func() time.Time
	dailyReport DailyReportNotifier
}

func NewLoginService(db *gorm.DB, dailyReport DailyReportNotifier) *ActivityService {
	return &ActivityService{db: db, now: time.Now, dailyReport: dailyReport}
}

func (service *ActivityService) Add(ctx context.Context, userID uuid.UUID, days []ActivityDay) error {
	if len(days) == 0 {
		return nil
	}

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		createdAt := service.now().UTC()

		for _, day := range days {
			activityDate := utcDate(day.Date)

			if !day.Active {
				if err := tx.
					Where("user_id = ? AND activity_date = ?", userID, activityDate).
					Delete(&models.UserLogin{}).Error; err != nil {
					return err
				}

				continue
			}

			activity := models.UserLogin{
				UserID:       userID,
				ActivityDate: activityDate,
				CreatedAt:    createdAt,
			}

			if err := tx.
				Clauses(clause.OnConflict{
					Columns: []clause.Column{
						{Name: "user_id"},
						{Name: "activity_date"},
					},
					DoNothing: true,
				}).
				Create(&activity).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("record user activities: %w", err)
	}

	service.dailyReport.Notify(userID)

	return nil
}

func (service *ActivityService) Get(ctx context.Context, userID uuid.UUID, date time.Time) (ActivityDay, error) {
	days, err := service.GetRange(ctx, userID, date, date)

	if err != nil {
		return ActivityDay{}, err
	}

	return days[0], nil
}

func (service *ActivityService) GetRange(ctx context.Context, userID uuid.UUID, dateFrom time.Time, dateTo time.Time) ([]ActivityDay, error) {
	dateFrom = utcDate(dateFrom)
	dateTo = utcDate(dateTo)
	daysCount := int(dateTo.Sub(dateFrom).Hours()/24) + 1

	if dateTo.Before(dateFrom) || daysCount < 1 || daysCount > maxActivityRangeDays {
		return nil, ErrInvalidActivityRange
	}

	var activities []models.UserLogin
	err := service.db.WithContext(ctx).
		Where("user_id = ? AND activity_date >= ? AND activity_date <= ?", userID, dateFrom, dateTo).
		Find(&activities).Error

	if err != nil {
		return nil, fmt.Errorf("get user activities: %w", err)
	}

	activeDates := make(map[string]struct{}, len(activities))

	for _, activity := range activities {
		activeDates[utcDate(activity.ActivityDate).Format(time.DateOnly)] = struct{}{}
	}

	return buildActivityDays(dateFrom, dateTo, activeDates), nil
}

func buildActivityDays(dateFrom time.Time, dateTo time.Time, activeDates map[string]struct{}) []ActivityDay {
	dateFrom = utcDate(dateFrom)
	dateTo = utcDate(dateTo)
	daysCount := int(dateTo.Sub(dateFrom).Hours()/24) + 1
	days := make([]ActivityDay, 0, daysCount)

	for date := dateFrom; !date.After(dateTo); date = date.AddDate(0, 0, 1) {
		_, active := activeDates[date.Format(time.DateOnly)]
		days = append(days, ActivityDay{Date: date, Active: active})
	}

	return days
}
