package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/auth"
	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/metrics"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/petstate"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/shop"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type petHandler struct {
	auth        *auth.Service
	pets        *pet.Service
	levelClaims *pet.LevelClaimsService
	shop        *shop.Service
	state       *petstate.Service
	metrics     *appmetrics.Metrics
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
	*petStateResponse
}

type petStateResponse struct {
	Happiness             float64 `json:"happiness"`
	CalculatedAt          string  `json:"calculatedAt"`
	DecaysToZeroAt        string  `json:"decaysToZeroAt"`
	StrokeNextAvailableAt *string `json:"strokeNextAvailableAt"`
	FeedNextAvailableAt   *string `json:"feedNextAvailableAt"`
	HappinessMultiplier   float64 `json:"happinessMultiplier"`
}

type petStateUpdatedEvent struct {
	Event string           `json:"event"`
	Data  petStateResponse `json:"data"`
}

type petProfileResponse struct {
	petProgressResponse
	BowlImageURL *string `json:"bowlImageUrl"`
	BedImageURL  *string `json:"bedImageUrl"`
}

type petCareRequest struct {
	Type petstate.CareType `json:"type"`
}

type petProgressUpdatedEvent struct {
	Event string               `json:"event"`
	Data  petProgressEventData `json:"data"`
}

type petProgressEventData struct {
	petProgressResponse
	BowlImageURL *string `json:"bowlImageUrl"`
	BedImageURL  *string `json:"bedImageUrl"`
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

	stateSnapshot, err := handler.state.Get(request.Context(), user.ID)

	if err != nil {
		writeJSON(response, http.StatusServiceUnavailable, map[string]string{"code": "PET_STATE_UNAVAILABLE", "message": "Pet state is temporarily unavailable"})

		return
	}

	activeImages, err := handler.shop.ActiveImageURLs(request.Context(), user.ID)

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load active pet items")

		return
	}

	result := petProfileResponse{
		petProgressResponse: responsePet(userPet),
		BowlImageURL:        activeImages.Bowl,
		BedImageURL:         activeImages.Bed,
	}
	result.petStateResponse = responsePetState(stateSnapshot)

	writeJSON(response, http.StatusOK, result)
}

func (handler *petHandler) care(response http.ResponseWriter, request *http.Request) {
	user, authenticated := handler.authenticate(response, request)

	if !authenticated {
		return
	}

	var body petCareRequest

	if decodeJSON(response, request, &body) != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"code": "INVALID_REQUEST", "message": "Invalid request body"})

		return
	}

	result, err := handler.state.Care(request.Context(), user.ID, body.Type, request.Header.Get("Idempotency-Key"))

	var cooldown *petstate.CooldownError

	switch {
	case errors.As(err, &cooldown):
		writeJSON(response, http.StatusConflict, map[string]string{"code": "PET_CARE_COOLDOWN", "message": err.Error(), "nextAvailableAt": cooldown.NextAvailableAt.UTC().Format(time.RFC3339Nano)})
	case errors.Is(err, petstate.ErrFull):
		writeJSON(response, http.StatusConflict, map[string]string{"code": "PET_HAPPINESS_FULL", "message": err.Error()})
	case errors.Is(err, petstate.ErrInvalidType):
		writeJSON(response, http.StatusBadRequest, map[string]string{"code": "INVALID_CARE_TYPE", "message": err.Error()})
	case errors.Is(err, petstate.ErrInvalidIdempotencyKey):
		writeJSON(response, http.StatusBadRequest, map[string]string{"code": "INVALID_IDEMPOTENCY_KEY", "message": err.Error()})
	case errors.Is(err, petstate.ErrIdempotencyConflict):
		writeJSON(response, http.StatusConflict, map[string]string{"code": "IDEMPOTENCY_KEY_CONFLICT", "message": err.Error()})
	case err != nil:
		writeJSON(response, http.StatusServiceUnavailable, map[string]string{"code": "PET_STATE_UNAVAILABLE", "message": "Pet state is temporarily unavailable"})
	default:
		writeJSON(response, http.StatusOK, responsePetState(result))
	}
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

	result := responsePet(userPet)

	if stateSnapshot, stateErr := handler.state.Get(request.Context(), user.ID); stateErr == nil {
		result.petStateResponse = responsePetState(stateSnapshot)
	}

	writeJSON(response, http.StatusOK, result)
}

func (handler *petHandler) addMVPLeaves(response http.ResponseWriter, request *http.Request) {
	user, isAuthenticated := handler.authenticate(response, request)

	if !isAuthenticated {

		return
	}

	progress, err := handler.pets.AddMVPLeaves(request.Context(), user.ID)

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not add MVP leaves")

		return
	}

	result := responsePetProgress(progress)

	if stateSnapshot, stateErr := handler.state.Get(request.Context(), user.ID); stateErr == nil {
		result.petStateResponse = responsePetState(stateSnapshot)
	}

	writeJSON(response, http.StatusOK, result)
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
	if handler.metrics != nil {
		handler.metrics.WebSockets.Inc()
		defer handler.metrics.WebSockets.Dec()
	}
	defer func() {
		closeWebSocket(connection)
	}()
	configureWebSocket(connection)

	updates, unsubscribe := handler.pets.Subscribe(user.ID)
	defer unsubscribe()

	stateUpdates, unsubscribeState := handler.state.Subscribe(user.ID)
	defer unsubscribeState()

	err = handler.writePetProgressEvent(request.Context(), connection, user.ID, pet.ProgressForPet(userPet, false))

	if err != nil {
		return
	}

	initialState, err := handler.state.Get(request.Context(), user.ID)

	if err != nil || writePetStateEvent(connection, initialState) != nil {
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
			if !isOpen || handler.writePetProgressEvent(request.Context(), connection, user.ID, update.Progress) != nil {
				return
			}
		case update, isOpen := <-stateUpdates:
			if !isOpen || writePetStateEvent(connection, update.State) != nil {
				return
			}
		case <-requests:
			currentPet, err := handler.pets.GetOrCreate(request.Context(), user.ID)
			if err != nil || handler.writePetProgressEvent(request.Context(), connection, user.ID, pet.ProgressForPet(currentPet, false)) != nil {
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

func (handler *petHandler) writePetProgressEvent(ctx context.Context, connection *websocket.Conn, userID uuid.UUID, progress pet.Progress) error {
	activeImages, err := handler.shop.ActiveImageURLs(ctx, userID)
	if err != nil {
		return err
	}

	return writeWebSocketJSON(connection, petProgressUpdatedEvent{
		Event: "PET_PROGRESS_UPDATED",
		Data:  responsePetProgressEventData(progress, activeImages),
	})
}

func responsePetProgressEventData(progress pet.Progress, activeImages shop.ActiveImageURLs) petProgressEventData {
	return petProgressEventData{
		petProgressResponse: responsePetProgress(progress),
		BowlImageURL:        activeImages.Bowl,
		BedImageURL:         activeImages.Bed,
	}
}

func responsePetState(snapshot petstate.Snapshot) *petStateResponse {
	formatOptional := func(value *time.Time) *string {
		if value == nil {
			return nil
		}

		formatted := value.UTC().Format(time.RFC3339Nano)

		return &formatted
	}

	return &petStateResponse{
		Happiness:             snapshot.Happiness,
		CalculatedAt:          snapshot.CalculatedAt.UTC().Format(time.RFC3339Nano),
		DecaysToZeroAt:        snapshot.DecaysToZeroAt.UTC().Format(time.RFC3339Nano),
		StrokeNextAvailableAt: formatOptional(snapshot.StrokeNextAvailableAt),
		FeedNextAvailableAt:   formatOptional(snapshot.FeedNextAvailableAt),
		HappinessMultiplier:   snapshot.HappinessMultiplier,
	}
}

func writePetStateEvent(connection *websocket.Conn, snapshot petstate.Snapshot) error {
	return writeWebSocketJSON(connection, petStateUpdatedEvent{Event: "PET_STATE_UPDATED", Data: *responsePetState(snapshot)})
}

func writeLevelRewardError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, map[string]string{"code": code, "message": message})
}
