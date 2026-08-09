package handlers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type petHandler struct {
	auth        *auth.Service
	pets        *pet.Service
	levelClaims *pet.LevelClaimsService
}

type petNameUpdateRequest struct {
	Name string `json:"name"`
}

type petProgressResponse struct {
	Name                  string `json:"name"`
	Level                 int    `json:"level"`
	Leaves                int64  `json:"leaves"`
	NextLevelTargetLeaves int64  `json:"nextLevelTargetLeaves"`
	ChestPrice            int64  `json:"chestPrice"`
	LevelUp               bool   `json:"levelUp"`
}

type petProgressUpdatedEvent struct {
	Event string              `json:"event"`
	Data  petProgressResponse `json:"data"`
}

type petSocketRequest struct {
	Action string `json:"action"`
	Type   string `json:"type"`
}

const getChestPriceAction = "GET_CHEST_PRICE"

type petLevelRewardResponse struct {
	ID          string                `json:"id"`
	Type        models.RewardCategory `json:"type"`
	Description string                `json:"description"`
}

type petLevelResponse struct {
	Level     int                      `json:"level"`
	Status    models.LevelRewardStatus `json:"status"`
	Reward    petLevelRewardResponse   `json:"reward"`
	ExpiresAt *string                  `json:"expiresAt"`
}

type petLevelsResponse struct {
	Levels []petLevelResponse `json:"levels"`
}

type claimLevelRewardResponse struct {
	Level  int                      `json:"level"`
	Status models.LevelRewardStatus `json:"status"`
}

var petWebSocketUpgrader = websocket.Upgrader{
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

func (handler *petHandler) get(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	userPet, err := handler.pets.GetOrCreate(request.Context(), user.ID)

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load pet")

		return
	}

	writeJSON(response, http.StatusOK, responsePet(userPet))
}

func (handler *petHandler) updateName(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	var nameUpdateRequest petNameUpdateRequest

	err := decodeJSON(response, request, &nameUpdateRequest)

	if err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")

		return
	}

	userPet, err := handler.pets.UpdateName(request.Context(), user.ID, nameUpdateRequest.Name)

	if errors.Is(err, pet.ErrInvalidName) {
		writeError(response, http.StatusBadRequest, err.Error())

		return
	}

	if errors.Is(err, pet.ErrPetNotFound) {
		writeError(response, http.StatusNotFound, "Pet not found")

		return
	}

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not update pet name")

		return
	}

	writeJSON(response, http.StatusOK, responsePet(userPet))
}

func (handler *petHandler) levels(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	levels, err := handler.levelClaims.GetLevels(request.Context(), user.ID)

	if errors.Is(err, pet.ErrPetNotFound) {
		writeLevelRewardError(response, http.StatusNotFound, "PET_NOT_FOUND", "Питомец пользователя не найден")

		return
	}

	if err != nil {
		writeLevelRewardError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось загрузить награды уровней")

		return
	}

	responseLevels := make([]petLevelResponse, 0, len(levels))

	for _, level := range levels {
		var expiresAt *string

		if level.ExpiresAt != nil {
			value := level.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z")
			expiresAt = &value
		}

		responseLevels = append(responseLevels, petLevelResponse{
			Level:  level.Level,
			Status: level.Status,
			Reward: petLevelRewardResponse{
				ID:          level.Reward.ID.String(),
				Type:        level.Reward.Category,
				Description: level.Reward.Description,
			},
			ExpiresAt: expiresAt,
		})
	}

	writeJSON(response, http.StatusOK, petLevelsResponse{Levels: responseLevels})
}

func (handler *petHandler) claimLevelReward(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {
		return
	}

	rewardID, err := uuid.Parse(request.PathValue("rewardId"))

	if err != nil {
		writeLevelRewardError(response, http.StatusBadRequest, "INVALID_LEVEL_REWARD_ID", "Передан некорректный идентификатор награды")

		return
	}

	result, err := handler.levelClaims.Claim(request.Context(), user.ID, rewardID)

	switch {
	case errors.Is(err, pet.ErrPetNotFound):
		writeLevelRewardError(response, http.StatusNotFound, "PET_NOT_FOUND", "Питомец пользователя не найден")
	case errors.Is(err, pet.ErrLevelRewardNotFound):
		writeLevelRewardError(response, http.StatusNotFound, "LEVEL_REWARD_NOT_FOUND", "Награда не принадлежит авторизованному пользователю")
	case errors.Is(err, pet.ErrLevelRewardLocked):
		writeLevelRewardError(response, http.StatusConflict, "LEVEL_REWARD_LOCKED", "Сначала необходимо достичь этого уровня")
	case errors.Is(err, pet.ErrLevelRewardFrozen):
		writeLevelRewardError(response, http.StatusConflict, "LEVEL_REWARD_FROZEN", "Срок получения награды истёк")
	case errors.Is(err, pet.ErrLevelRewardAlreadyClaimed):
		writeLevelRewardError(response, http.StatusConflict, "LEVEL_REWARD_ALREADY_CLAIMED", "Награда за этот уровень уже забрана")
	case err != nil:
		writeLevelRewardError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось получить награду")
	default:
		writeJSON(response, http.StatusOK, claimLevelRewardResponse{Level: result.Level, Status: result.Status})
	}
}

func (handler *petHandler) ws(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), websocketToken(request))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	userPet, err := handler.pets.GetOrCreate(request.Context(), user.ID)

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load pet")

		return
	}

	connection, err := petWebSocketUpgrader.Upgrade(response, request, nil)

	if err != nil {
		return
	}
	defer func() {
		closeWebSocket(connection)
	}()
	configureWebSocket(connection)

	updates, unsubscribe := handler.pets.Subscribe(user.ID)
	defer unsubscribe()

	err = writePetProgressEvent(connection, pet.ProgressForPet(userPet, false))

	if err != nil {
		return
	}

	done := make(chan struct{})
	requests := make(chan petSocketRequest, 1)
	pingTicker := time.NewTicker(websocketPingPeriod)
	defer pingTicker.Stop()

	go func() {
		defer close(done)

		for {
			_, payload, err := connection.ReadMessage()

			if err != nil {
				return
			}

			var socketRequest petSocketRequest

			if json.Unmarshal(payload, &socketRequest) == nil {
				action := socketRequest.Action

				if action == "" {
					action = socketRequest.Type
				}

				if !strings.EqualFold(action, getChestPriceAction) {
					continue
				}

				select {
				case requests <- socketRequest:
				default:
				}
			}
		}
	}()

	for {
		select {
		case update, isOpen := <-updates:
			if !isOpen || writePetProgressEvent(connection, update.Progress) != nil {
				return
			}
		case <-requests:
			currentPet, err := handler.pets.GetOrCreate(request.Context(), user.ID)
			if err != nil || writePetProgressEvent(connection, pet.ProgressForPet(currentPet, false)) != nil {
				return
			}
		case <-done:
			return
		case <-pingTicker.C:
			if err := writeWebSocketPing(connection); err != nil {
				return
			}
		}
	}
}

func (handler *petHandler) authenticate(response http.ResponseWriter, request *http.Request) (models.User, bool) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return models.User{}, false
	}

	return user, true
}

func websocketToken(request *http.Request) string {
	if token := request.Header.Get("Authorization"); token != "" {
		return token
	}

	return request.URL.Query().Get("token")
}

func responsePet(userPet models.Pet) petProgressResponse {
	return responsePetProgress(pet.ProgressForPet(userPet, false))
}

func responsePetProgress(progress pet.Progress) petProgressResponse {
	chestPrice := progress.ChestPrice

	if chestPrice == 0 {
		chestPrice = models.ChestOpeningLeavesCost
	}

	return petProgressResponse{
		Name:                  progress.Name,
		Level:                 progress.Level,
		Leaves:                progress.Leaves,
		NextLevelTargetLeaves: progress.NextLevelTargetLeaves,
		ChestPrice:            chestPrice,
		LevelUp:               progress.LevelUp,
	}
}

func writePetProgressEvent(connection *websocket.Conn, progress pet.Progress) error {
	return writeWebSocketJSON(connection, petProgressUpdatedEvent{
		Event: "PET_PROGRESS_UPDATED",
		Data:  responsePetProgress(progress),
	})
}

func writeLevelRewardError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, map[string]string{"code": code, "message": message})
}
