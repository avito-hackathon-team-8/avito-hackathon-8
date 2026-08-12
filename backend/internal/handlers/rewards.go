package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/google/uuid"
)

type rewardHandler struct {
	auth    *auth.Service
	rewards *rewards.Service
}

type rewardResponse struct {
	ID         string               `json:"id"`
	Title      string               `json:"title"`
	Source     models.RewardSource  `json:"source"`
	Active     bool                 `json:"active"`
	Status     string               `json:"status"`
	ExpiresAt  time.Time            `json:"expiresAt"`
	RedeemedAt *time.Time           `json:"redeemedAt"`
	ItemType   *models.ShopItemType `json:"itemType"`
}

type rewardDetailsResponse = rewardResponse

type rewardGroupResponse struct {
	Category models.RewardCategory `json:"category"`
	Items    []rewardResponse      `json:"items"`
}

type rewardsResponse struct {
	Groups []rewardGroupResponse `json:"groups"`
}

type redeemRewardResponse struct {
	ID         string    `json:"id"`
	RedeemedAt time.Time `json:"redeemedAt"`
}

func (handler *rewardHandler) list(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	userRewards, err := handler.rewards.List(request.Context(), user.ID)

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load rewards")

		return
	}

	groupsByCategory := make(map[models.RewardCategory][]rewardResponse)

	for _, reward := range userRewards {
		status := handler.rewards.Status(reward)
		groupsByCategory[reward.Category] = append(groupsByCategory[reward.Category], responseReward(reward, status))
	}

	groups := make([]rewardGroupResponse, 0, len(groupsByCategory))

	for _, category := range rewards.CategoryOrder {
		items := groupsByCategory[category]

		if len(items) == 0 {
			continue
		}

		groups = append(groups, rewardGroupResponse{
			Category: category,
			Items:    items,
		})
	}

	writeJSON(response, http.StatusOK, rewardsResponse{Groups: groups})
}

func (handler *rewardHandler) get(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	rewardID, err := uuid.Parse(request.PathValue("rewardId"))

	if err != nil {
		writeError(response, http.StatusBadRequest, "Invalid reward ID")

		return
	}

	reward, err := handler.rewards.Get(request.Context(), user.ID, rewardID)

	if errors.Is(err, rewards.ErrRewardNotFound) {
		writeError(response, http.StatusNotFound, "Reward not found")

		return
	}

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not load reward")

		return
	}

	writeJSON(response, http.StatusOK, responseRewardDetails(reward, handler.rewards.Status(reward)))
}

func (handler *rewardHandler) redeem(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	rewardID, err := uuid.Parse(request.PathValue("rewardId"))

	if err != nil {
		writeError(response, http.StatusBadRequest, "Invalid reward ID")

		return
	}

	reward, err := handler.rewards.Redeem(request.Context(), user.ID, rewardID)

	if errors.Is(err, rewards.ErrRewardNotFound) {
		writeError(response, http.StatusNotFound, "Active reward not found")

		return
	}

	if err != nil {
		writeError(response, http.StatusInternalServerError, "Could not redeem reward")

		return
	}

	writeJSON(response, http.StatusOK, redeemRewardResponse{
		ID:         reward.ID.String(),
		RedeemedAt: *reward.RedeemedAt,
	})
}

func responseReward(reward models.Reward, status string) rewardResponse {
	return rewardResponse{
		ID:         reward.ID.String(),
		Title:      reward.Title,
		Source:     reward.Source,
		Active:     status == rewards.StatusActive,
		Status:     status,
		ExpiresAt:  reward.ExpiresAt,
		RedeemedAt: reward.RedeemedAt,
		ItemType:   reward.ItemType,
	}
}

func responseRewardDetails(reward models.Reward, status string) rewardDetailsResponse {
	return rewardDetailsResponse{
		ID:         reward.ID.String(),
		Title:      reward.Title,
		Source:     reward.Source,
		Active:     status == rewards.StatusActive,
		Status:     status,
		ExpiresAt:  reward.ExpiresAt,
		RedeemedAt: reward.RedeemedAt,
		ItemType:   reward.ItemType,
	}
}

func writeAuthenticationError(response http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrUnauthorized) {
		writeError(response, http.StatusUnauthorized, err.Error())

		return
	}

	writeError(response, http.StatusInternalServerError, "Could not load the user")
}
