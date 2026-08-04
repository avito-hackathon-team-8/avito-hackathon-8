package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
)

type authHandler struct {
	service *auth.Service
}

type otpRequest struct {
	Email string `json:"email"`
}

type otpVerification struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type userResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

func NewRouter(service *auth.Service) http.Handler {
	handler := &authHandler{service: service}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("POST /api/app/auth/request-otp", handler.requestOTP)
	mux.HandleFunc("POST /api/app/auth/verify-otp", handler.verifyOTP)
	mux.HandleFunc("GET /api/app/auth/me", handler.me)

	return mux
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

	user, token, err := handler.service.VerifyOTP(request.Context(), body.Email, body.Code)

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
	user, err := handler.service.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			writeError(response, http.StatusUnauthorized, err.Error())

			return
		}

		writeError(response, http.StatusInternalServerError, "Could not load the user")

		return
	}

	writeJSON(response, http.StatusOK, responseUser(user))
}

func responseUser(user models.User) userResponse {
	return userResponse{ID: user.ID.String(), Email: user.Email, Verified: user.Verified}
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 1<<20))
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func writeError(response http.ResponseWriter, status int, message string) {
	writeJSON(response, status, map[string]string{"message": message})
}

func writeJSON(response http.ResponseWriter, status int, body any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}
