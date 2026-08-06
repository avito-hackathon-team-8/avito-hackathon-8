package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/weekly_login"
)

type weeklyLoginHandler struct {
	auth        *auth.Service
	activity    weekly_login.ActivityProvider
	weeklyLogin *weekly_login.Service
}

type weeklyLoginDayResponse struct {
	Weekday      int    `json:"weekday"`
	Date         string `json:"date"`
	Status       string `json:"status"`
	RewardLeaves int    `json:"rewardLeaves"`
	ClaimID      string `json:"claimId,omitempty"`
}

type weeklyLoginResponse struct {
	ClaimedDaysCount int                      `json:"claimedDaysCount"`
	Claims           []weeklyLoginDayResponse `json:"claims"`
}

type weeklyLoginClaimRequest struct {
	Date string `json:"date"`
}

type weeklyLoginClaim struct {
	ID           string `json:"id"`
	Weekday      int    `json:"weekday"`
	Date         string `json:"date"`
	Status       string `json:"status"`
	RewardLeaves int    `json:"rewardLeaves"`
}

type weeklyLoginClaimResponse struct {
	Claim weeklyLoginClaim `json:"claim"`
}

type dailyActivityRequest struct {
	Days *[]dailyActivityDayRequest `json:"days"`
}

type dailyActivityDayRequest struct {
	Date   string `json:"date"`
	Active *bool  `json:"active"`
}

func (handler *weeklyLoginHandler) addActivity(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	var body dailyActivityRequest

	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")

		return
	}

	if body.Days == nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")

		return
	}

	days := make([]weekly_login.ActivityDay, 0, len(*body.Days))

	for _, item := range *body.Days {
		if item.Active == nil {
			writeError(response, http.StatusBadRequest, "Invalid request body")

			return
		}

		date, err := time.Parse(time.DateOnly, item.Date)

		if err != nil {
			writeError(response, http.StatusBadRequest, "Invalid activity date")

			return
		}

		days = append(days, weekly_login.ActivityDay{Date: date, Active: *item.Active})
	}

	if err := handler.activity.Add(request.Context(), user.ID, days); err != nil {
		writeError(response, http.StatusInternalServerError, "Could not record daily activity")

		return
	}

	response.WriteHeader(http.StatusNoContent)
}

func (handler *weeklyLoginHandler) get(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	weeklyLogin, err := handler.weeklyLogin.Get(request.Context(), user.ID)

	if errors.Is(err, weekly_login.ErrUserNotFound) {
		writeError(response, http.StatusNotFound, err.Error())

		return
	}

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load weekly login")

		return
	}

	writeJSON(response, http.StatusOK, responseWeeklyLogin(weeklyLogin))
}

func (handler *weeklyLoginHandler) claim(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	var body weeklyLoginClaimRequest

	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")

		return
	}

	date, err := time.Parse(time.DateOnly, body.Date)

	if err != nil {
		writeError(response, http.StatusBadRequest, "Invalid date")

		return
	}

	claim, err := handler.weeklyLogin.Claim(request.Context(), user.ID, date)

	if errors.Is(err, weekly_login.ErrAlreadyClaimed) {
		writeJSON(response, http.StatusConflict, map[string]string{
			"code":    "WEEKLY_LOGIN_REWARD_ALREADY_CLAIMED",
			"message": err.Error(),
		})

		return
	}

	if errors.Is(err, weekly_login.ErrActivityNotConfirmed) {
		writeError(response, http.StatusBadRequest, err.Error())

		return
	}

	if errors.Is(err, weekly_login.ErrUserNotFound) {
		writeError(response, http.StatusNotFound, err.Error())

		return
	}

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not claim weekly login reward")

		return
	}

	writeJSON(response, http.StatusOK, weeklyLoginClaimResponse{Claim: responseWeeklyLoginClaim(claim)})
}

func responseWeeklyLogin(weeklyLogin weekly_login.CurrentWeek) weeklyLoginResponse {
	claims := make([]weeklyLoginDayResponse, 0, len(weeklyLogin.Claims))

	for _, claim := range weeklyLogin.Claims {
		claimID := ""

		if claim.ClaimID != nil {
			claimID = claim.ClaimID.String()
		}

		claims = append(claims, weeklyLoginDayResponse{
			Weekday:      claim.Weekday,
			Date:         claim.Date,
			Status:       string(claim.Status),
			RewardLeaves: claim.RewardLeaves,
			ClaimID:      claimID,
		})
	}

	return weeklyLoginResponse{
		ClaimedDaysCount: weeklyLogin.ClaimedDaysCount,
		Claims:           claims,
	}
}

func responseWeeklyLoginClaim(claim models.WeeklyLoginClaim) weeklyLoginClaim {
	return weeklyLoginClaim{
		ID:           claim.ID.String(),
		Weekday:      isoWeekday(claim.ClaimDate),
		Date:         claim.ClaimDate.Format(time.DateOnly),
		Status:       string(models.DayStatusClaimed),
		RewardLeaves: claim.RewardLeaves,
	}
}

func isoWeekday(date time.Time) int {
	return (int(date.UTC().Weekday())+6)%7 + 1
}
