package health

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobState struct {
	LastRun   time.Time `json:"lastRun,omitempty"`
	LastError string    `json:"lastError,omitempty"`
}

type Status struct {
	mu   sync.RWMutex
	jobs map[string]JobState
}

func NewStatus() *Status {
	return &Status{jobs: make(map[string]JobState)}
}

func (status *Status) Record(name string, at time.Time, err error) {
	state := JobState{LastRun: at.UTC()}

	if err != nil {
		state.LastError = err.Error()
	}

	status.mu.Lock()
	status.jobs[name] = state
	status.mu.Unlock()
}

func (status *Status) Snapshot() map[string]JobState {
	status.mu.RLock()
	defer status.mu.RUnlock()

	result := make(map[string]JobState, len(status.jobs))

	for name, state := range status.jobs {
		result[name] = state
	}

	return result
}

type TaskEnsurer interface {
	Ensure(context.Context, uuid.UUID) error
}

func NewHandler(status *Status, db *gorm.DB, internalToken string, tasks TaskEnsurer) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(response http.ResponseWriter, request *http.Request) {
		sqlDB, err := db.DB()

		if err == nil {
			err = sqlDB.PingContext(request.Context())
		}

		code := http.StatusOK
		state := "ready"

		if err != nil {
			code = http.StatusServiceUnavailable
			state = "unavailable"
		}

		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(code)

		_ = json.NewEncoder(response).Encode(map[string]any{
			"service": "puppeteer",
			"status":  state,
			"jobs":    status.Snapshot(),
			"time":    time.Now().UTC(),
		})
	})

	mux.HandleFunc("POST /internal/v1/users/{userId}/daily-tasks/ensure", func(response http.ResponseWriter, request *http.Request) {
		provided := request.Header.Get("X-Service-Token")

		if !serviceTokenEqual(provided, internalToken) {
			writeError(response, http.StatusUnauthorized, "invalid service token")

			return
		}

		userID, err := uuid.Parse(request.PathValue("userId"))

		if err != nil {
			writeError(response, http.StatusBadRequest, "invalid user id")

			return
		}

		if err := tasks.Ensure(request.Context(), userID); err != nil {
			writeError(response, http.StatusInternalServerError, "could not ensure daily tasks")

			return
		}

		response.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func serviceTokenEqual(provided, expected string) bool {
	providedHash := sha256.Sum256([]byte(provided))
	expectedHash := sha256.Sum256([]byte(expected))

	return subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1
}

func writeError(response http.ResponseWriter, status int, message string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]string{"message": message})
}
