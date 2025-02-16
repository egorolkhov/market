package sendCoin

import (
	"avito/internal/cache"
	"avito/internal/models"
	"avito/internal/storage"
	"avito/internal/utils"
	"avito/internal/utils/jwtToken"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type Storage interface {
	Send(ctx context.Context, uuid string, toUser string, amount int) error
	GetUuidByUsername(ctx context.Context, username string) (string, error)
}

type SendController struct {
	Storage Storage
	Lfu     *cache.LFUCache
}

type SendRequest struct {
	ToUser string `json:"toUser" validate:"required"`
	Amount int    `json:"amount" validate:"required,ne=0"`
}

func (sc *SendController) SendCoin(w http.ResponseWriter, r *http.Request) {
	var req SendRequest

	validate := validator.New()
	err := json.NewDecoder(r.Body).Decode(&req)
	defer r.Body.Close()
	errValidate := validate.Struct(req)
	if err != nil || errValidate != nil {
		response := models.ErrorResponse{Errors: "the request parameters are incorrect."}
		utils.JsonResponse(w, http.StatusBadRequest, response)
		return
	}

	cookie := r.Header.Get("Authorization")
	uuid := jwtToken.GetUserID(cookie)
	if uuid == "" {
		response := models.ErrorResponse{Errors: "internal server error."}
		utils.JsonResponse(w, http.StatusInternalServerError, response)
		return
	}
	err = sc.Storage.Send(r.Context(), uuid, req.ToUser, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotEnoughBalance):
			response := models.ErrorResponse{Errors: storage.ErrNotEnoughBalance.Error()}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		case errors.Is(err, storage.ErrSendingToYourself):
			response := models.ErrorResponse{Errors: storage.ErrSendingToYourself.Error()}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		case errors.Is(err, storage.ErrUserNotFound):
			response := models.ErrorResponse{Errors: storage.ErrUserNotFound.Error()}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		default:
			response := models.ErrorResponse{Errors: "error sending tokens."}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		}
	}

	receiverUuid, err := sc.Storage.GetUuidByUsername(r.Context(), req.ToUser)
	if err != nil {
		sc.Lfu.ClearCache()
	}
	sc.Lfu.Delete(receiverUuid)
	sc.Lfu.Delete(uuid)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
