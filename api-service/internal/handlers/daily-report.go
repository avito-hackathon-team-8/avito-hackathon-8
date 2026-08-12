package handlers

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/daily_report"
	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/metrics"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/gorilla/websocket"
)

type dailyReportRewardResponse struct {
	RewardID   string                `json:"rewardId"`
	Type       models.RewardCategory `json:"type"`
	Title      string                `json:"title"`
	ExpiresAt  string                `json:"expiresAt"`
	ReceivedAt time.Time             `json:"receivedAt"`
}

type dailyReportTaskResponse struct {
	TaskID       string          `json:"taskId"`
	Type         models.TaskType `json:"type"`
	Description  string          `json:"description"`
	RewardLeaves int             `json:"rewardLeaves"`
	CompletedAt  time.Time       `json:"completedAt"`
}

type dailyReportLevelUpResponse struct {
	FromLevel  int       `json:"fromLevel"`
	ToLevel    int       `json:"toLevel"`
	OccurredAt time.Time `json:"occurredAt"`
}

type dailyReportResponse struct {
	LeavesEarnedToday int                         `json:"leavesEarnedToday"`
	Date              string                      `json:"date"`
	Rewards           []dailyReportRewardResponse `json:"rewards"`
	LevelUp           *dailyReportLevelUpResponse `json:"levelUp"`
	Tasks             []dailyReportTaskResponse   `json:"tasks"`
	VisitedToday      bool                        `json:"visitedToday"`
	UpdatedAt         time.Time                   `json:"updatedAt"`
}

type dailyReportUpdatedEvent struct {
	Event string              `json:"event"`
	Data  dailyReportResponse `json:"data"`
}

type dailyReportHandler struct {
	auth        *auth.Service
	dailyReport *daily_report.Service
	metrics     *appmetrics.Metrics
}

var dailyReportsWebSocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(request *http.Request) bool {
		origin := request.Header.Get("Origin")

		if origin == "" {
			return true
		}

		parsed, err := url.Parse(origin)

		if err != nil {
			return false
		}

		requestHost := request.Host

		if host, _, splitErr := net.SplitHostPort(requestHost); splitErr == nil {
			requestHost = host
		}

		return strings.EqualFold(parsed.Hostname(), requestHost)
	},
}

func (h *dailyReportHandler) get(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.Authenticate(r.Context(), r.Header.Get("Authorization"))

	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())

		return
	}

	dailyReport, err := h.dailyReport.Get(r.Context(), user.ID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())

		return
	}

	writeJSON(w, http.StatusOK, responseDailyReport(dailyReport))
}

func (h *dailyReportHandler) ws(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.Authenticate(r.Context(), websocketToken(r))

	if err != nil {
		writeAuthenticationError(w, err)

		return
	}

	updates, unsubscribe := h.dailyReport.Subscribe(user.ID)
	defer unsubscribe()

	report, err := h.dailyReport.Get(r.Context(), user.ID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())

		return
	}

	ws, err := dailyReportsWebSocketUpgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	h.metrics.WebSockets.Inc()
	defer h.metrics.WebSockets.Dec()

	defer func() {
		closeWebSocket(ws)
	}()

	configureWebSocket(ws)

	if err := writeDailyReportEvent(ws, report); err != nil {
		return
	}

	done := make(chan struct{})
	pingTicker := time.NewTicker(websocketPingPeriod)
	defer pingTicker.Stop()

	go func() {
		defer close(done)

		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				return
			}
		}
	}()

	dayTimer := time.NewTimer(untilNextUTCDay(time.Now()))
	defer dayTimer.Stop()

	writeCurrentReport := func() error {
		report, err := h.dailyReport.Get(r.Context(), user.ID)

		if err != nil {
			return err
		}

		return writeDailyReportEvent(ws, report)
	}

	for {
		select {
		case _, isOpen := <-updates:
			if !isOpen || writeCurrentReport() != nil {
				return
			}
		case <-dayTimer.C:
			if writeCurrentReport() != nil {
				return
			}
			dayTimer.Reset(untilNextUTCDay(time.Now()))
		case <-done:
			return
		case <-pingTicker.C:
			if err := writeWebSocketPing(ws); err != nil {
				return
			}
		}
	}
}

func writeDailyReportEvent(connection *websocket.Conn, report daily_report.DailyReport) error {
	return writeWebSocketJSON(connection, dailyReportUpdatedEvent{
		Event: "DAILY_REPORT_UPDATED",
		Data:  responseDailyReport(report),
	})
}

func untilNextUTCDay(now time.Time) time.Duration {
	now = now.UTC()
	nextDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)

	return nextDay.Sub(now)
}

func responseDailyReport(report daily_report.DailyReport) dailyReportResponse {
	rewards := make([]dailyReportRewardResponse, 0, len(report.Rewards))

	for _, reward := range report.Rewards {
		rewards = append(rewards, dailyReportRewardResponse{
			RewardID:   reward.ID.String(),
			Type:       reward.Type,
			Title:      reward.Title,
			ExpiresAt:  reward.ExpiresAt.UTC().Format(time.DateOnly),
			ReceivedAt: reward.ReceivedAt.UTC(),
		})
	}

	tasks := make([]dailyReportTaskResponse, 0, len(report.Tasks))

	for _, task := range report.Tasks {
		tasks = append(tasks, dailyReportTaskResponse{
			TaskID:       task.ID.String(),
			Type:         task.Type,
			Description:  task.Description,
			RewardLeaves: task.RewardLeaves,
			CompletedAt:  task.CompletedAt.UTC(),
		})
	}

	var levelUp *dailyReportLevelUpResponse

	if report.LevelUp != nil {
		levelUp = &dailyReportLevelUpResponse{
			FromLevel:  report.LevelUp.FromLevel,
			ToLevel:    report.LevelUp.ToLevel,
			OccurredAt: report.LevelUp.OccurredAt.UTC(),
		}
	}

	return dailyReportResponse{
		LeavesEarnedToday: report.LeavesEarnedToday,
		Date:              report.Date,
		Rewards:           rewards,
		LevelUp:           levelUp,
		Tasks:             tasks,
		VisitedToday:      report.VisitedToday,
		UpdatedAt:         report.UpdatedAt.UTC(),
	}
}
