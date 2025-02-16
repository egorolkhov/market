package buy

import (
	"avito/internal/cache"
	"avito/internal/models"
	"avito/internal/storage"
	"avito/internal/utils"
	"avito/internal/utils/jwtToken"
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"net/http"
)

type Storage interface {
	Buy(ctx context.Context, uuid string, item string) error
}

type BuyParams struct {
	Item string `schema:"item" validate:"required"`
}

type BuyController struct {
	Storage Storage
	Lfu     *cache.LFUCache
}

func (bc *BuyController) Buy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	item, ok := vars["item"]
	if !ok {
		response := models.ErrorResponse{Errors: "the request parameters are incorrect"}
		utils.JsonResponse(w, http.StatusBadRequest, response)
		return
	}

	decoder := schema.NewDecoder()
	validate := validator.New()

	var params BuyParams
	err := decoder.Decode(&params, r.URL.Query())
	params.Item = item
	errValidate := validate.Struct(params)
	if err != nil || errValidate != nil {
		response := models.ErrorResponse{Errors: "the request parameters are incorrect"}
		utils.JsonResponse(w, http.StatusBadRequest, response)
		return
	}

	cookie := r.Header.Get("Authorization")
	uuid := jwtToken.GetUserID(cookie)
	if uuid == "" {
		response := models.ErrorResponse{Errors: "internal server error"}
		utils.JsonResponse(w, http.StatusInternalServerError, response)
		return
	}
	err = bc.Storage.Buy(r.Context(), uuid, item)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrItemNotFound):
			response := models.ErrorResponse{Errors: storage.ErrItemNotFound.Error()}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		case errors.Is(err, storage.ErrNotEnoughBalance):
			response := models.ErrorResponse{Errors: storage.ErrNotEnoughBalance.Error()}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		default:
			response := models.ErrorResponse{Errors: "error buying item"}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		}
	}

	bc.Lfu.Delete(uuid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
