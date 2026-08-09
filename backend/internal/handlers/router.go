package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/chest"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/daily_report"
	activityevents "github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/events"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/tasks"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/weekly_login"
	"gorm.io/gorm"
)

type authHandler struct {
	service *auth.Service
	db      *gorm.DB
}

type otpRequest struct {
	Email string `json:"email"`
}

type otpVerification struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type userResponse struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email"`
	Verified    bool                   `json:"verified"`
	Leaderboard *leaderboardMeResponse `json:"leaderboard,omitempty"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(db *gorm.DB, authService *auth.Service, rewardService *rewards.Service,
	taskService *tasks.Service, petService *pet.Service,
	levelClaimsService *pet.LevelClaimsService, weeklyLoginService *weekly_login.Service,
	eventService *activityevents.Service, dailyReportService *daily_report.Service,
	internalToken string, chestService *chest.Service,
) http.Handler {
	handler := &authHandler{service: authService, db: db}
	rewardHandler := &rewardHandler{auth: authService, rewards: rewardService}
	taskHandler := &taskHandler{
		auth: authService, tasks: taskService, pets: petService,
	}
	petHandler := &petHandler{
		auth:        authService,
		pets:        petService,
		levelClaims: levelClaimsService,
	}
	chestHandler := &chestHandler{
		auth:    authService,
		chests:  chestService,
		rewards: rewardService,
	}
	weeklyLoginHandler := &weeklyLoginHandler{
		auth:        authService,
		weeklyLogin: weeklyLoginService,
		pets:        petService,
	}
	leaderboardHandler := &leaderboardHandler{auth: authService, db: db}
	dailyReportHandler := &dailyReportHandler{auth: authService, dailyReport: dailyReportService}
	internalEvents := &internalEventHandler{token: internalToken, events: eventService}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("POST /api/app/auth/request-otp", handler.requestOTP)
	mux.HandleFunc("POST /api/app/auth/verify-otp", handler.verifyOTP)
	mux.HandleFunc("GET /api/app/auth/me", handler.me)

	mux.HandleFunc("GET /api/app/rewards", rewardHandler.list)
	mux.HandleFunc("GET /api/app/rewards/{rewardId}", rewardHandler.get)
	mux.HandleFunc("POST /api/app/rewards/{rewardId}/redeem", rewardHandler.redeem)

	mux.HandleFunc("GET /api/v1/pet", petHandler.get)
	mux.HandleFunc("PATCH /api/v1/pet", petHandler.updateName)
	mux.HandleFunc("GET /api/v1/pet/ws", petHandler.ws)
	mux.HandleFunc("GET /api/v1/pet/levels", petHandler.levels)
	mux.HandleFunc("POST /api/v1/pet/level-rewards/{rewardId}/claim", petHandler.claimLevelReward)

	mux.HandleFunc("POST /api/v1/pet/chests/open", chestHandler.open)

	mux.HandleFunc("GET /api/v1/tasks", taskHandler.list)
	mux.HandleFunc("GET /api/v1/tasks/progress", taskHandler.progress)
	mux.HandleFunc("POST /api/v1/tasks/record", taskHandler.record)
	mux.HandleFunc("POST /api/v1/tasks/{taskId}/claim", taskHandler.claim)

	mux.HandleFunc("POST /api/v1/weekly-login/activity", weeklyLoginHandler.addActivity)
	mux.HandleFunc("GET /api/v1/weekly-login", weeklyLoginHandler.get)
	mux.HandleFunc("POST /api/v1/weekly-login/claim", weeklyLoginHandler.claim)
	mux.HandleFunc("GET /api/v1/daily-report", dailyReportHandler.get)
	mux.HandleFunc("GET /api/v1/daily-report/ws", dailyReportHandler.ws)
	mux.HandleFunc("GET /api/v1/leaderboard", leaderboardHandler.list)
	mux.HandleFunc("POST /api/internal/v1/users/{userId}/events", internalEvents.record)

	return withCORS(mux)
}

func health(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{
		"service": "go-api",
		"status":  "ready",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (handler *authHandler) requestOTP(response http.ResponseWriter, request *http.Request) {
	var body otpRequest

	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")

		return
	}

	if err := handler.service.RequestOTP(request.Context(), body.Email); err != nil {
		if errors.Is(err, auth.ErrInvalidEmail) {
			writeError(response, http.StatusBadRequest, err.Error())

			return
		}

		writeError(response, http.StatusInternalServerError, "Could not send the sign-in code")

		return
	}

	writeJSON(response, http.StatusOK, map[string]bool{"sent": true})
}

func (handler *authHandler) verifyOTP(response http.ResponseWriter, request *http.Request) {
	var body otpVerification

	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")

		return
	}

	user, token, err := handler.service.VerifyOTP(
		request.Context(),
		body.Email,
		body.Code,
	)

	if err != nil {
		if errors.Is(err, auth.ErrInvalidOTP) {
			writeError(response, http.StatusBadRequest, err.Error())

			return
		}

		writeError(response, http.StatusInternalServerError, "Could not verify the sign-in code")

		return
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"token":  token,
		"record": responseUser(user),
	})
}

func (handler *authHandler) me(response http.ResponseWriter, request *http.Request) {
	user, err := handler.service.Authenticate(
		request.Context(),
		request.Header.Get("Authorization"),
	)

	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			writeError(response, http.StatusUnauthorized, err.Error())

			return
		}

		writeError(response, http.StatusInternalServerError, "Could not load the user")

		return
	}

	leaderboard, err := loadLeaderboardMe(request.Context(), handler.db, user, time.Now().UTC())

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load leaderboard position")

		return
	}

	currentUser := responseUser(user)
	currentUser.Leaderboard = leaderboard

	writeJSON(response, http.StatusOK, currentUser)
}

func responseUser(user models.User) userResponse {
	return userResponse{
		ID:       user.ID.String(),
		Email:    user.Email,
		Verified: user.Verified,
	}
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) error {
	decoder := json.NewDecoder(
		http.MaxBytesReader(response, request.Body, 1<<20),
	)
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func writeError(response http.ResponseWriter, status int, message string) {
	writeJSON(response, status, errorResponse{
		Code:    errorCode(status),
		Message: message,
	})
}

func errorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "INVALID_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	default:
		return "INTERNAL_ERROR"
	}
}

func writeJSON(response http.ResponseWriter, status int, body any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}
