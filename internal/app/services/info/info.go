package info

import (
	"avito/internal/cache"
	"avito/internal/models"
	"avito/internal/utils"
	"avito/internal/utils/jwtToken"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type Storage interface {
	GetInfo(ctx context.Context, uuid string) (*models.Info, error)
}

type InfoController struct {
	Storage Storage
	Lfu     *cache.LFUCache
}

func (ic *InfoController) Info(w http.ResponseWriter, r *http.Request) {
	cookie := r.Header.Get("Authorization")
	uuid := jwtToken.GetUserID(cookie)
	if uuid == "" {
		response := models.ErrorResponse{Errors: "internal server error."}
		utils.JsonResponse(w, http.StatusInternalServerError, response)
		return
	}

	key := uuid
	if val, ok := ic.Lfu.Get(key); ok {
		var info models.Info
		cachedBody := val.(string)
		err := json.Unmarshal([]byte(cachedBody), &info)
		if err == nil {
			utils.JsonResponse(w, http.StatusOK, info)
			return
		}
		log.Println("Can't get cache info")
	}

	info, err := ic.Storage.GetInfo(r.Context(), uuid)
	if err != nil {
		response := models.ErrorResponse{Errors: "error finding uuid"}
		utils.JsonResponse(w, http.StatusInternalServerError, response)
		return
	}

	infoByte, err := json.Marshal(info)
	if err != nil {
		log.Println("Can't cache info")
	} else {
		ic.Lfu.Set(key, string(infoByte))
	}

	utils.JsonResponse(w, http.StatusOK, info)
}
