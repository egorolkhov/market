package info

import (
	"avito/internal/cache"
	"avito/internal/models"
	"avito/internal/utils/jwtToken"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct {
	GetInfoFunc func(ctx context.Context, uuid string) (*models.Info, error)
}

func (m *mockStorage) GetInfo(ctx context.Context, uuid string) (*models.Info, error) {
	return m.GetInfoFunc(ctx, uuid)
}

func TestInfoController_Info(t *testing.T) {
	lfu := cache.NewLFUCache(10)
	controller := &InfoController{
		Storage: &mockStorage{},
		Lfu:     lfu,
	}

	t.Run("info found in cache -> 200", func(t *testing.T) {
		uuid := "uuid-123"
		info := &models.Info{
			Coins: 100,
			CoinsHistory: models.History{
				Received: []models.Transaction{{FromUser: "user1", Amount: 50}},
				Sent:     []models.Transaction{{ToUser: "user2", Amount: 20}},
			},
			Inventory: []models.Item{{Type: "t-shirt", Quantity: 1}},
		}

		infoByte, _ := json.Marshal(info)
		lfu.Set(uuid, string(infoByte))

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		token, _ := jwtToken.BuidToken("uuid-123")

		w := httptest.NewRecorder()
		req.Header.Set("Authorization", token)
		controller.Info(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("info not found in DB -> 500", func(t *testing.T) {
		mockSt := &mockStorage{
			GetInfoFunc: func(ctx context.Context, uuid string) (*models.Info, error) {
				return nil, errors.New("error finding uuid")
			},
		}
		controller.Storage = mockSt

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		token, _ := jwtToken.BuidToken("uuid-456")
		req.Header.Set("Authorization", token)

		w := httptest.NewRecorder()
		controller.Info(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("info retrieved successfully -> 200", func(t *testing.T) {
		mockSt := &mockStorage{
			GetInfoFunc: func(ctx context.Context, uuid string) (*models.Info, error) {
				return &models.Info{
					Coins:     150,
					Inventory: []models.Item{{Type: "pen", Quantity: 1}},
				}, nil
			},
		}
		controller.Storage = mockSt

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		token, _ := jwtToken.BuidToken("uuid-123")
		req.Header.Set("Authorization", token)

		w := httptest.NewRecorder()
		controller.Info(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
