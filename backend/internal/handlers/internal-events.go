package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	activityevents "github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/events"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/tasks"
	"github.com/google/uuid"
)

type internalEventHandler struct {
	token  string
	events *activityevents.Service
}

type internalEventBatchRequest struct {
	Events []internalEventRequest `json:"events"`
}

type internalEventRequest struct {
	EventID    string `json:"eventId"`
	Type       string `json:"type"`
	Count      int    `json:"count"`
	OccurredAt string `json:"occurredAt"`
}

func (handler *internalEventHandler) record(response http.ResponseWriter, request *http.Request) {
	provided := request.Header.Get("X-Service-Token")

	if !serviceTokenEqual(provided, handler.token) {
		writeJSON(response, http.StatusUnauthorized, taskErrorResponse{Code: "INVALID_SERVICE_TOKEN", Message: "Invalid service token"})

		return
	}

	userID, err := uuid.Parse(request.PathValue("userId"))

	if err != nil {
		writeTaskError(response, http.StatusBadRequest, "INVALID_USER_ID", "Некорректный userId.")

		return
	}

	var body internalEventBatchRequest

	if err := decodeJSON(response, request, &body); err != nil || len(body.Events) == 0 {
		writeTaskError(response, http.StatusBadRequest, "INVALID_REQUEST", "Некорректное тело запроса.")

		return
	}

	batch := make([]activityevents.Event, 0, len(body.Events))

	for _, item := range body.Events {
		eventID, idErr := uuid.Parse(item.EventID)
		occurredAt, timeErr := time.Parse(time.RFC3339, item.OccurredAt)

		if idErr != nil || timeErr != nil {
			writeTaskError(response, http.StatusBadRequest, "INVALID_EVENT", "Некорректное событие.")

			return
		}

		batch = append(batch, activityevents.Event{ID: eventID, Type: item.Type, Count: item.Count, OccurredAt: occurredAt})
	}

	err = handler.events.Record(request.Context(), userID, batch)

	switch {
	case errors.Is(err, activityevents.ErrUserNotFound):
		writeTaskError(response, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден.")
	case errors.Is(err, activityevents.ErrEventConflict):
		writeTaskError(response, http.StatusConflict, "EVENT_ID_CONFLICT", err.Error())
	case errors.Is(err, activityevents.ErrInvalidEvent), errors.Is(err, activityevents.ErrEventOutsideTime):
		writeTaskError(response, http.StatusBadRequest, "INVALID_EVENT", err.Error())
	case errors.Is(err, tasks.ErrTaskLocked):
		writeTaskError(response, http.StatusForbidden, "TASK_LOCKED", "Задание недоступно.")
	case errors.Is(err, tasks.ErrTaskNotFound), errors.Is(err, tasks.ErrTasksNotReady):
		writeTaskError(response, http.StatusServiceUnavailable, "TASKS_NOT_READY", "Задания ещё назначаются.")
	case err != nil:
		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")
	default:
		response.WriteHeader(http.StatusNoContent)
	}
}

func serviceTokenEqual(provided, expected string) bool {
	providedHash := sha256.Sum256([]byte(provided))
	expectedHash := sha256.Sum256([]byte(expected))

	return subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1
}
