package handlers

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/daily_report"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
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
	TaskID        string          `json:"taskId"`
	Type          models.TaskType `json:"type"`
	Description   string          `json:"description"`
	RewardLeaves  int             `json:"rewardLeaves"`
	RewardClaimed bool            `json:"rewardClaimed"`
	CompletedAt   time.Time       `json:"completedAt"`
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
}

// TODO: Вынести дубликат
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

func (h *dailyReportHandler) get(response http.ResponseWriter, request *http.Request) {
	user, err := h.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))
	if err != nil {
		writeError(response, http.StatusUnauthorized, err.Error())
	}

	dailyReport, err := h.dailyReport.Get(request.Context(), user.ID)
	if err != nil {
		writeError(response, http.StatusInternalServerError, err.Error())
	}

	writeJSON(response, http.StatusOK, responseDailyReport(dailyReport))
}

func (h *dailyReportHandler) ws(response http.ResponseWriter, request *http.Request) {
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
			TaskID:        task.ID.String(),
			Type:          task.Type,
			Description:   task.Description,
			RewardLeaves:  task.RewardLeaves,
			RewardClaimed: task.RewardClaimed,
			CompletedAt:   task.CompletedAt.UTC(),
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
