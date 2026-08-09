package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/tasks"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type taskHandler struct {
	auth  *auth.Service
	pets  *pet.Service
	tasks *tasks.Service
}

type dailyTaskResponse struct {
	TaskID        string            `json:"taskId"`
	Slot          int               `json:"slot"`
	Type          models.TaskType   `json:"type"`
	Description   string            `json:"description"`
	CurrentCount  int               `json:"currentCount"`
	TargetCount   int               `json:"targetCount"`
	RewardLeaves  int               `json:"rewardLeaves"`
	RequiredLevel int               `json:"requiredLevel"`
	Status        models.TaskStatus `json:"status"`
}

type dailyTasksResponse struct {
	Tasks []dailyTaskResponse `json:"tasks"`
}

type dailyTasksProgressResponse struct {
	CompletedCount int `json:"completedCount"`
	TotalCount     int `json:"totalCount"`
}

type dailyTaskRecordRequest struct {
	Events []EventItem `json:"events"`
}

type EventItem struct {
	TaskID string          `json:"taskId,omitempty"`
	Type   models.TaskType `json:"type"`
	Count  int             `json:"count"`
}

type dailyTaskClaimResponse struct {
	TaskID       string            `json:"taskId"`
	RewardLeaves int               `json:"rewardLeaves"`
	Status       models.TaskStatus `json:"status"`
}

type taskErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (handler *taskHandler) list(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	userLevel, hasPetLevel := handler.petLevel(response, request, user.ID)

	if !hasPetLevel {
		return
	}

	if _, err := handler.tasks.List(request.Context(), user.ID, userLevel); err != nil {
		if errors.Is(err, tasks.ErrTasksNotReady) {
			writeTaskError(response, http.StatusServiceUnavailable, "TASKS_NOT_READY", "Задания ещё назначаются. Повторите запрос.")

			return
		}

		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return
	}

	if err := handler.tasks.AutoCompleteFirstTasks(request.Context(), user.ID); err != nil {
		if errors.Is(err, tasks.ErrTasksNotReady) {
			writeTaskError(response, http.StatusServiceUnavailable, "TASKS_NOT_READY", "Задания ещё назначаются. Повторите запрос.")

			return
		}

		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return
	}

	dailyTasks, err := handler.tasks.List(request.Context(), user.ID, userLevel)

	if err != nil {
		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return
	}

	responseTasks := make([]dailyTaskResponse, 0, len(dailyTasks))

	for _, task := range dailyTasks {
		responseTasks = append(responseTasks, responseDailyTask(task))
	}

	writeJSON(response, http.StatusOK, dailyTasksResponse{Tasks: responseTasks})
}

func (handler *taskHandler) progress(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	userLevel, hasPetLevel := handler.petLevel(response, request, user.ID)

	if !hasPetLevel {
		return
	}

	if _, err := handler.tasks.Progress(request.Context(), user.ID, userLevel); err != nil {
		if errors.Is(err, tasks.ErrTasksNotReady) {
			writeTaskError(response, http.StatusServiceUnavailable, "TASKS_NOT_READY", "Задания ещё назначаются. Повторите запрос.")

			return
		}

		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return
	}

	if err := handler.tasks.AutoCompleteFirstTasks(request.Context(), user.ID); err != nil {
		if errors.Is(err, tasks.ErrTasksNotReady) {
			writeTaskError(response, http.StatusServiceUnavailable, "TASKS_NOT_READY", "Задания ещё назначаются. Повторите запрос.")

			return
		}

		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return
	}

	progress, err := handler.tasks.Progress(request.Context(), user.ID, userLevel)

	if err != nil {
		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return
	}

	writeJSON(response, http.StatusOK, dailyTasksProgressResponse{
		CompletedCount: progress.CompletedCount,
		TotalCount:     progress.TotalCount,
	})
}

func (handler *taskHandler) record(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	var recordRequest dailyTaskRecordRequest

	err := decodeJSON(response, request, &recordRequest)

	if err != nil {
		writeTaskError(response, http.StatusBadRequest, "INVALID_REQUEST", "Некорректное тело запроса.")

		return
	}

	userLevel, hasPetLevel := handler.petLevel(response, request, user.ID)

	if !hasPetLevel {
		return
	}

	events := make([]tasks.Event, 0, len(recordRequest.Events))

	for _, event := range recordRequest.Events {
		count := event.Count

		if count < 1 {
			count = 1
		}

		var taskID uuid.UUID

		if event.TaskID != "" {
			taskID, err = uuid.Parse(event.TaskID)

			if err != nil {
				writeTaskError(response, http.StatusBadRequest, "INVALID_REQUEST", "Некорректный taskId.")

				return
			}
		}

		events = append(events, tasks.Event{TaskID: taskID, Type: event.Type, Count: count})
	}

	err = handler.tasks.RecordEvents(request.Context(), user.ID, events, userLevel)

	switch {
	case errors.Is(err, tasks.ErrInvalidTaskType):
		writeTaskError(response, http.StatusBadRequest, "INVALID_TASK_TYPE", "Передан неизвестный тип задания.")
	case errors.Is(err, tasks.ErrTaskLocked):
		writeTaskError(response, http.StatusForbidden, "TASK_LOCKED", "Одно из заданий недоступно на текущем уровне питомца.")
	case errors.Is(err, tasks.ErrTaskNotFound):
		writeTaskError(response, http.StatusNotFound, "TASK_NOT_FOUND", "Одно из заданий текущего дня не найдено.")
	case err != nil:
		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")
	default:
		response.WriteHeader(http.StatusNoContent)
	}
}

func (handler *taskHandler) claim(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	taskID, err := uuid.Parse(request.PathValue("taskId"))

	if err != nil {
		writeTaskError(response, http.StatusNotFound, "TASK_NOT_FOUND", "Задание текущего дня не найдено.")

		return
	}

	userLevel, hasPetLevel := handler.petLevel(response, request, user.ID)

	if !hasPetLevel {
		return
	}

	var progress pet.Progress

	claimResult, err := handler.tasks.ClaimWithReward(request.Context(), user.ID, taskID, userLevel, func(tx *gorm.DB, amount int) error {
		var err error

		progress, err = handler.pets.CreditTx(tx, pet.Credit{
			UserID: user.ID, Amount: int64(amount), Reason: models.LeafReasonTaskReward,
			OperationKey: fmt.Sprintf("task:%s", taskID),
		})

		return err
	})

	switch {
	case errors.Is(err, tasks.ErrTaskLocked):
		writeTaskError(response, http.StatusForbidden, "TASK_LOCKED", "Задание недоступно на текущем уровне питомца.")
	case errors.Is(err, tasks.ErrTaskNotFound):
		writeTaskError(response, http.StatusNotFound, "TASK_NOT_FOUND", "Задание текущего дня не найдено.")
	case errors.Is(err, tasks.ErrTaskNotCompleted):
		writeTaskError(response, http.StatusConflict, "TASK_NOT_COMPLETED", "Сначала выполните условие задания.")
	case errors.Is(err, tasks.ErrRewardAlreadyClaimed):
		writeTaskError(response, http.StatusConflict, "TASK_REWARD_ALREADY_CLAIMED", "Награда за это задание уже получена.")
	case err != nil:
		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")
	default:
		handler.pets.PublishProgress(user.ID, progress)
		writeJSON(response, http.StatusOK, dailyTaskClaimResponse{
			TaskID:       claimResult.TaskID.String(),
			RewardLeaves: claimResult.RewardLeaves,
			Status:       claimResult.Status,
		})
	}
}

func (handler *taskHandler) authenticate(response http.ResponseWriter, request *http.Request) (models.User, bool) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if errors.Is(err, auth.ErrUnauthorized) {
		writeTaskError(response, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется аутентификация")

		return models.User{}, false
	}

	if err != nil {
		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return models.User{}, false
	}

	return user, true
}

func (handler *taskHandler) petLevel(response http.ResponseWriter, request *http.Request, userID uuid.UUID) (int, bool) {
	pet, err := handler.pets.GetOrCreate(request.Context(), userID)

	if err != nil {
		writeTaskError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return 0, false
	}

	return pet.Level, true
}

func responseDailyTask(task tasks.DailyTask) dailyTaskResponse {
	return dailyTaskResponse{
		TaskID:        task.ID.String(),
		Slot:          task.Slot,
		Type:          task.Type,
		Description:   task.Description,
		CurrentCount:  task.CurrentCount,
		TargetCount:   task.TargetCount,
		RewardLeaves:  task.RewardLeaves,
		RequiredLevel: task.RequiredLevel,
		Status:        task.Status,
	}
}

func writeTaskError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, taskErrorResponse{Code: code, Message: message})
}
