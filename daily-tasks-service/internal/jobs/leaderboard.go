package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CalculateLeaderboard struct {
	Rewards  map[int]LeaderboardReward
	Observer WorkObserver
}

func (CalculateLeaderboard) Name() string { return "calculate-leaderboard" }

func (job CalculateLeaderboard) Run(ctx context.Context, db *gorm.DB, now time.Time) error {
	now = now.UTC()
	current := startOfMonth(now)
	previous := current.AddDate(0, -1, 0)

	if err := job.finalizePeriod(ctx, db, previous, current, now); err != nil {
		return err
	}

	if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&models.LeaderboardSeason{PeriodStart: current}).Error; err != nil {
		return fmt.Errorf("create current leaderboard season: %w", err)
	}

	if err := calculatePeriod(ctx, db, current, current.AddDate(0, 1, 0), now, job.Observer); err != nil {
		return fmt.Errorf("calculate current leaderboard: %w", err)
	}

	return nil
}

func (job CalculateLeaderboard) finalizePeriod(ctx context.Context, db *gorm.DB, periodStart, periodEnd, now time.Time) error {
	season := models.LeaderboardSeason{PeriodStart: periodStart}

	if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&season).Error; err != nil {
		return fmt.Errorf("create leaderboard season: %w", err)
	}

	if err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&season, "period_start = ?", periodStart).Error; err != nil {
		return fmt.Errorf("lock leaderboard season: %w", err)
	}

	if season.FinalizedAt != nil {
		return nil
	}

	if err := calculatePeriod(ctx, db, periodStart, periodEnd, now, job.Observer); err != nil {
		return fmt.Errorf("calculate final leaderboard: %w", err)
	}

	var winners []models.LeaderboardEntry

	if err := db.WithContext(ctx).Where("period_start = ? AND rank <= 3", periodStart).Order("rank ASC").Find(&winners).Error; err != nil {
		return fmt.Errorf("load leaderboard winners: %w", err)
	}

	for _, winner := range winners {
		definition, ok := job.Rewards[int(winner.Rank)]

		if !ok {
			return fmt.Errorf("leaderboard reward rank %d is not configured", winner.Rank)
		}

		reference := fmt.Sprintf("%s:rank:%d", periodStart.Format("2006-01"), winner.Rank)
		reward := models.Reward{
			UserID: winner.UserID, Title: definition.Title, Category: definition.Category,
			Source: "LEADERBOARD", SourceReference: &reference, ExpiresAt: now.Add(definition.ValidFor), CreatedAt: now,
		}

		if err := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "source"}, {Name: "source_reference"}}, DoNothing: true,
		}).Create(&reward).Error; err != nil {
			return fmt.Errorf("grant leaderboard reward: %w", err)
		}
	}

	finalizedAt := now

	if err := db.WithContext(ctx).Model(&season).Update("finalized_at", finalizedAt).Error; err != nil {
		return fmt.Errorf("finalize leaderboard season: %w", err)
	}

	return nil
}

func calculatePeriod(ctx context.Context, db *gorm.DB, periodStart, periodEnd, calculatedAt time.Time, observers ...WorkObserver) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("period_start = ?", periodStart).Delete(&models.LeaderboardEntry{}).Error; err != nil {
			return fmt.Errorf("clear leaderboard: %w", err)
		}

		result := tx.Exec(`
		INSERT INTO leaderboard_entries (period_start, user_id, leaves, rank, calculated_at)
		SELECT ?, user_id, leaves,
			ROW_NUMBER() OVER (ORDER BY leaves DESC, created_at ASC, user_id ASC), ?
		FROM (
			SELECT users.id AS user_id, users.created_at,
				COALESCE(totals.leaves, 0) AS leaves
			FROM users
			LEFT JOIN leaderboard_totals totals ON totals.user_id = users.id AND totals.period_start = ?
			WHERE users.verified = TRUE AND users.created_at < ?
		) ranked_totals`, periodStart, calculatedAt.UTC(), periodStart, periodEnd)

		if result.Error != nil {
			return fmt.Errorf("insert leaderboard: %w", result.Error)
		}

		if len(observers) > 0 && observers[0] != nil {
			observers[0].AddRows("calculate-leaderboard", int(result.RowsAffected))
		}
		return nil
	})
}

func startOfMonth(at time.Time) time.Time {
	at = at.UTC()

	return time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, time.UTC)
}
