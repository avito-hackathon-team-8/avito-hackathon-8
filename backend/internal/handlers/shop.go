package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/shop"
)

type shopHandler struct {
	auth *auth.Service
	shop *shop.Service
}

type shopItemResponse struct {
	ID            string              `json:"id"`
	Type          models.ShopItemType `json:"type"`
	Status        shop.ItemStatus     `json:"status"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	RequiredLevel int                 `json:"requiredLevel"`
	PriceLeaves   int64               `json:"priceLeaves"`
	DurationDays  int                 `json:"durationDays"`
}

type shopResponse struct {
	Items []shopItemResponse `json:"items"`
}

type shopPurchaseRequest struct {
	ConfirmReplacement bool `json:"confirmReplacement"`
}

type shopErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (handler *shopHandler) list(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	items, err := handler.shop.List(request.Context(), user.ID)
	if err != nil {
		writeShopError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось загрузить магазин")
		return
	}

	responseItems := make([]shopItemResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, responseShopItem(item))
	}

	writeJSON(response, http.StatusOK, shopResponse{Items: responseItems})
}

func (handler *shopHandler) purchase(response http.ResponseWriter, request *http.Request) {
	user, err := handler.auth.Authenticate(request.Context(), request.Header.Get("Authorization"))

	if err != nil {
		writeAuthenticationError(response, err)

		return
	}

	var body shopPurchaseRequest
	if err := decodeJSON(response, request, &body); err != nil && !errors.Is(err, io.EOF) {
		writeShopError(response, http.StatusBadRequest, "INVALID_REQUEST", "Некорректное тело запроса")
		return
	}

	err = handler.shop.Purchase(request.Context(), user.ID, shop.Purchase{
		ItemID:             request.PathValue("itemId"),
		ConfirmReplacement: body.ConfirmReplacement,
	})

	switch {
	case errors.Is(err, shop.ErrItemNotFound):
		writeShopError(response, http.StatusNotFound, "SHOP_ITEM_NOT_FOUND", "Предмет магазина не найден")
	case errors.Is(err, shop.ErrPetNotFound):
		writeShopError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось купить предмет")
	case errors.Is(err, shop.ErrLevelRequired):
		writeShopError(response, http.StatusConflict, "SHOP_LEVEL_REQUIRED", "Для покупки требуется более высокий уровень питомца")
	case errors.Is(err, shop.ErrInsufficientLeaves):
		writeShopError(response, http.StatusConflict, "INSUFFICIENT_LEAVES", "Недостаточно листьев для покупки предмета")
	case errors.Is(err, shop.ErrReplacementConfirmation):
		writeShopError(response, http.StatusConflict, "SHOP_REPLACEMENT_CONFIRMATION_REQUIRED", "Подтвердите замену активного предмета этого типа")
	case err != nil:
		writeShopError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось купить предмет")
	default:
		response.WriteHeader(http.StatusNoContent)
	}
}

func responseShopItem(item shop.Item) shopItemResponse {
	return shopItemResponse{
		ID:            item.ID,
		Type:          item.Type,
		Status:        item.Status,
		Title:         item.Title,
		Description:   item.Description,
		RequiredLevel: item.RequiredLevel,
		PriceLeaves:   item.PriceLeaves,
		DurationDays:  item.DurationDays,
	}
}

func writeShopError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, shopErrorResponse{Code: code, Message: message})
}
