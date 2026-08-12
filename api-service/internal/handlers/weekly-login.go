package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/weekly_login"
)

type weeklyLoginHandler struct {
	auth        *auth.Service
	weeklyLogin *weekly_login.Service
	pets        *pet.Service
}

type weeklyLoginDayResponse struct {
	Weekday             int     `json:"weekday"`
	Date                string  `json:"date"`
	Status              string  `json:"status"`
	RewardLeaves        int     `json:"rewardLeaves"`
	BaseRewardLeaves    int     `json:"baseRewardLeaves"`
	HappinessMultiplier float64 `json:"happinessMultiplier"`
	ClaimID             string  `json:"claimId,omitempty"`
}

type weeklyLoginResponse struct {
	ClaimedDaysCount int                      `json:"claimedDaysCount"`
	Claims           []weeklyLoginDayResponse `json:"claims"`
}

type weeklyLoginClaim struct {
	ID                  string  `json:"id"`
	Weekday             int     `json:"weekday"`
	Date                string  `json:"date"`
	Status              string  `json:"status"`
	RewardLeaves        int     `json:"rewardLeaves"`
	BaseRewardLeaves    int     `json:"baseRewardLeaves"`
	HappinessMultiplier float64 `json:"happinessMultiplier"`
}

type weeklyLoginClaimResponse struct {
	Claim weeklyLoginClaim `json:"claim"`
}

func (handler *weeklyLoginHandler) addActivity(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	if err := handler.weeklyLogin.RecordToday(request.Context(), user.ID); err != nil {
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

	claimResult, err := handler.weeklyLogin.Claim(request.Context(), user.ID)

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

	handler.pets.PublishProgress(user.ID, claimResult.Progress)

	writeJSON(response, http.StatusOK, weeklyLoginClaimResponse{Claim: responseWeeklyLoginClaim(claimResult.Claim)})
}

func responseWeeklyLogin(weeklyLogin weekly_login.CurrentWeek) weeklyLoginResponse {
	claims := make([]weeklyLoginDayResponse, 0, len(weeklyLogin.Claims))

	for _, claim := range weeklyLogin.Claims {
		claimID := ""

		if claim.ClaimID != nil {
			claimID = claim.ClaimID.String()
		}

		claims = append(claims, weeklyLoginDayResponse{
			Weekday:             claim.Weekday,
			Date:                claim.Date,
			Status:              string(claim.Status),
			RewardLeaves:        claim.RewardLeaves,
			BaseRewardLeaves:    claim.BaseRewardLeaves,
			HappinessMultiplier: claim.HappinessMultiplier,
			ClaimID:             claimID,
		})
	}

	return weeklyLoginResponse{
		ClaimedDaysCount: weeklyLogin.ClaimedDaysCount,
		Claims:           claims,
	}
}

func responseWeeklyLoginClaim(claim models.WeeklyLoginClaim) weeklyLoginClaim {
	return weeklyLoginClaim{
		ID:                  claim.ID.String(),
		Weekday:             isoWeekday(claim.ClaimDate),
		Date:                claim.ClaimDate.Format(time.DateOnly),
		Status:              string(models.DayStatusClaimed),
		RewardLeaves:        claim.RewardLeaves,
		BaseRewardLeaves:    claim.BaseRewardLeaves,
		HappinessMultiplier: valueOrOne(claim.HappinessMultiplier),
	}
}

func valueOrOne(value *float64) float64 {
	if value == nil {
		return 1
	}

	return *value
}

func isoWeekday(date time.Time) int {
	return (int(date.UTC().Weekday())+6)%7 + 1
}
