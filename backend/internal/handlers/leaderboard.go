package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type leaderboardHandler struct {
	auth *auth.Service
	db   *gorm.DB
}

type leaderboardUser struct {
	PlayerID string `json:"playerId"`
	Nickname string `json:"nickname"`
	Position int64  `json:"position"`
	Leaves   int64  `json:"leaves"`
}

type leaderboardResponse struct {
	Period            leaderboardPeriod `json:"period"`
	CalculatedAt      time.Time         `json:"calculatedAt"`
	NextCalculationAt time.Time         `json:"nextCalculationAt"`
	Items             []leaderboardUser `json:"items"`
}

type leaderboardMeResponse struct {
	Period            leaderboardPeriod `json:"period"`
	CalculatedAt      time.Time         `json:"calculatedAt"`
	NextCalculationAt time.Time         `json:"nextCalculationAt"`
	Player            leaderboardMe     `json:"player"`
}

type leaderboardPeriod struct {
	Key     string    `json:"key"`
	StartAt time.Time `json:"startAt"`
	EndAt   time.Time `json:"endAt"`
}

type leaderboardMe struct {
	PlayerID string `json:"playerId"`
	Nickname string `json:"nickname"`
	Position int64  `json:"position"`
	Leaves   int64  `json:"leaves"`
	IsTop10  bool   `json:"isTop10"`
}

type leaderboardRow struct {
	PlayerID     uuid.UUID
	Nickname     string
	Position     int64
	Leaves       int64
	CalculatedAt time.Time
}

func (handler *leaderboardHandler) list(response http.ResponseWriter, request *http.Request) {
	if _, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization")); err != nil {
		writeAuthenticationError(response, err)
		return
	}
	now := time.Now().UTC()
	period := startOfMonth(now)
	var rows []leaderboardRow
	if err := handler.db.WithContext(request.Context()).Table("leaderboard_entries AS entries").Select(`
		entries.user_id AS player_id, split_part(users.email, '@', 1) AS nickname,
		entries.rank AS position, entries.leaves, entries.calculated_at`).
		Joins("JOIN users ON users.id = entries.user_id").Where("entries.period_start = ?", period).
		Order("entries.rank ASC").Limit(10).Scan(&rows).Error; err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load leaderboard")
		return
	}
	calculatedAt, err := handler.snapshotCalculatedAt(request.Context(), period)
	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load leaderboard")
		return
	}
	if calculatedAt.IsZero() {
		writeError(response, http.StatusServiceUnavailable, "Leaderboard is not ready")
		return
	}
	items := make([]leaderboardUser, 0, len(rows))
	for _, row := range rows {
		items = append(items, leaderboardUser{PlayerID: row.PlayerID.String(), Nickname: row.Nickname, Position: row.Position, Leaves: row.Leaves})
	}
	writeJSON(response, http.StatusOK, leaderboardResponse{
		Period: makePeriod(period), CalculatedAt: calculatedAt,
		NextCalculationAt: nextMidnight(now), Items: items,
	})
}

func loadLeaderboardMe(ctx context.Context, db *gorm.DB, user models.User, now time.Time) (*leaderboardMeResponse, error) {
	handler := &leaderboardHandler{db: db}
	period := startOfMonth(now)
	calculatedAt, err := handler.snapshotCalculatedAt(ctx, period)

	if err != nil {
		return nil, err
	}

	if calculatedAt.IsZero() {
		return nil, nil
	}

	var row leaderboardRow

	err = db.WithContext(ctx).Table("leaderboard_entries AS entries").Select(`
		entries.user_id AS player_id, split_part(users.email, '@', 1) AS nickname,
		entries.rank AS position, entries.leaves, entries.calculated_at`).
		Joins("JOIN users ON users.id = entries.user_id").
		Where("entries.period_start = ? AND entries.user_id = ?", period, user.ID).Take(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		var count int64

		if err := db.WithContext(ctx).Table("leaderboard_entries").Where("period_start = ?", period).Count(&count).Error; err != nil {
			return nil, err
		}

		row = leaderboardRow{
			PlayerID: user.ID, Nickname: nicknameFromEmail(user.Email),
			Position: count + 1, Leaves: 0, CalculatedAt: calculatedAt,
		}
	} else if err != nil {
		return nil, err
	}

	return &leaderboardMeResponse{
		Period: makePeriod(period), CalculatedAt: calculatedAt,
		NextCalculationAt: nextMidnight(now),
		Player: leaderboardMe{
			PlayerID: row.PlayerID.String(), Nickname: row.Nickname,
			Position: row.Position, Leaves: row.Leaves, IsTop10: row.Position <= 10,
		},
	}, nil
}

func (handler *leaderboardHandler) snapshotCalculatedAt(ctx context.Context, period time.Time) (time.Time, error) {
	var calculatedAt sql.NullTime

	err := handler.db.WithContext(ctx).Raw(
		"SELECT MAX(calculated_at) FROM leaderboard_entries WHERE period_start = ?", period).Scan(&calculatedAt).Error

	if err != nil {
		return time.Time{}, err
	}

	if calculatedAt.Valid {
		return calculatedAt.Time.UTC(), nil
	}

	err = handler.db.WithContext(ctx).Raw(`
		SELECT MAX(ran_at) FROM job_runs
		WHERE job_name = ? AND ran_at >= ? AND ran_at < ?`, "calculate-leaderboard", period, period.AddDate(0, 1, 0)).Scan(&calculatedAt).Error

	if err != nil || !calculatedAt.Valid {
		return time.Time{}, err
	}

	return calculatedAt.Time.UTC(), nil
}

func nicknameFromEmail(email string) string {
	for index, value := range email {
		if value == '@' {
			return email[:index]
		}
	}

	return email
}

func makePeriod(start time.Time) leaderboardPeriod {
	return leaderboardPeriod{Key: start.Format("2006-01"), StartAt: start, EndAt: start.AddDate(0, 1, 0)}
}

func startOfMonth(at time.Time) time.Time {
	at = at.UTC()

	return time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func nextMidnight(at time.Time) time.Time {
	at = at.UTC()

	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
}
