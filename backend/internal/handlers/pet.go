package handlers

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/gorilla/websocket"
)

type petHandler struct {
	auth *auth.Service
	pets *pet.Service
}

type petResponse struct {
	Name   string `json:"name"`
	Level  int    `json:"level"`
	Leaves int64  `json:"leaves"`
}

type petNameUpdateRequest struct {
	Name string `json:"name"`
}

type petProgressResponse struct {
	Name                  string `json:"name"`
	Level                 int    `json:"level"`
	Leaves                int64  `json:"leaves"`
	NextLevelTargetLeaves int64  `json:"nextLevelTargetLeaves"`
	LevelUp               bool   `json:"levelUp"`
}

type petProgressUpdatedEvent struct {
	Event string              `json:"event"`
	Data  petProgressResponse `json:"data"`
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
	user, ok := handler.authenticate(response, request)
	if !ok {
		return
	}

	currentPet, err := handler.pets.GetOrCreate(request.Context(), user.ID)
	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load pet")
		return
	}

	writeJSON(response, http.StatusOK, responsePet(currentPet))
}

func (handler *petHandler) updateName(response http.ResponseWriter, request *http.Request) {
	user, ok := handler.authenticate(response, request)
	if !ok {
		return
	}

	var body petNameUpdateRequest
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")
		return
	}

	currentPet, err := handler.pets.UpdateName(request.Context(), user.ID, body.Name)
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

	writeJSON(response, http.StatusOK, responsePet(currentPet))
}

func (handler *petHandler) ws(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), websocketToken(request))
	if err != nil {
		writeAuthenticationError(response, err)
		return
	}

	currentPet, err := handler.pets.GetOrCreate(request.Context(), user.ID)
	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load pet")
		return
	}

	connection, err := petWebSocketUpgrader.Upgrade(response, request, nil)
	if err != nil {
		return
	}
	defer connection.Close()

	updates, unsubscribe := handler.pets.Subscribe(user.ID)
	defer unsubscribe()

	if err := writePetProgressEvent(connection, pet.ProgressForPet(currentPet, false)); err != nil {
		return
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := connection.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case update, ok := <-updates:
			if !ok || writePetProgressEvent(connection, update.Progress) != nil {
				return
			}
		case <-done:
			return
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

func responsePet(currentPet models.Pet) petResponse {
	return petResponse{
		Name:   currentPet.Name,
		Level:  currentPet.Level,
		Leaves: currentPet.Leaves,
	}
}

func writePetProgressEvent(connection *websocket.Conn, progress pet.Progress) error {
	return connection.WriteJSON(petProgressUpdatedEvent{
		Event: "PET_PROGRESS_UPDATED",
		Data: petProgressResponse{
			Name:                  progress.Name,
			Level:                 progress.Level,
			Leaves:                progress.Leaves,
			NextLevelTargetLeaves: progress.NextLevelTargetLeaves,
			LevelUp:               progress.LevelUp,
		},
	})
}
