package handlers

import (
	"errors"
	"net/http"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/chest"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
)

type chestHandler struct {
	auth    *auth.Service
	chests  *chest.Service
	rewards *rewards.Service
}

type chestErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (handler *chestHandler) open(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)
	if !isAuthenticated {
		return
	}

	reward, err := handler.chests.Open(request.Context(), user.ID)

	switch {
	case errors.Is(err, chest.ErrPetNotFound):
		writeChestError(response, http.StatusNotFound, "PET_NOT_FOUND", "Питомец пользователя не найден")
	case errors.Is(err, chest.ErrChestLevelRequired):
		writeChestError(response, http.StatusConflict, "CHEST_LEVEL_REQUIRED", "Сундуки доступны с 10-го уровня питомца")
	case errors.Is(err, chest.ErrInsufficientLeaves):
		writeChestError(response, http.StatusConflict, "INSUFFICIENT_LEAVES", "Для открытия сундука требуется 200 листьев")
	case err != nil:
		writeChestError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось открыть сундук")
	default:
		writeJSON(response, http.StatusOK, responseReward(reward, handler.rewards.Status(reward)))
	}
}

func (handler *chestHandler) authenticate(response http.ResponseWriter, request *http.Request) (models.User, bool) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if errors.Is(err, auth.ErrUnauthorized) {
		writeChestError(response, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется аутентификация")

		return models.User{}, false
	}

	if err != nil {
		writeChestError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")

		return models.User{}, false
	}

	return user, true
}

func writeChestError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, chestErrorResponse{Code: code, Message: message})
}
