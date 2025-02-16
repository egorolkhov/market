package buy

import (
	"avito/internal/cache"
	"avito/internal/storage"
	"avito/internal/utils/jwtToken"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct {
	BuyFunc func(ctx context.Context, uuid string, item string) error
}

func (m *mockStorage) Buy(ctx context.Context, uuid, item string) error {
	return m.BuyFunc(ctx, uuid, item)
}

func TestBuyController_Buy(t *testing.T) {
	lfu := cache.NewLFUCache(10)
	controller := &BuyController{
		Storage: &mockStorage{},
		Lfu:     lfu,
	}

	t.Run("missing item in URL vars", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/buy/", nil)
		w := httptest.NewRecorder()
		controller.Buy(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "should get status bad request")
	})

	t.Run("buy success -> 200", func(t *testing.T) {
		mockSt := &mockStorage{
			BuyFunc: func(ctx context.Context, uuid, item string) error {
				assert.Equal(t, "user-uuid-123", uuid)
				assert.Equal(t, "t-shirt", item)
				return nil
			},
		}
		controller.Storage = mockSt

		req := httptest.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
		vars := map[string]string{
			"item": "t-shirt",
		}
		req = mux.SetURLVars(req, vars)

		w := httptest.NewRecorder()
		token, _ := jwtToken.BuidToken("user-uuid-123")
		req.Header.Set("Authorization", token)

		controller.Buy(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("no JWT -> 500", func(t *testing.T) {
		mockSt := &mockStorage{
			BuyFunc: func(ctx context.Context, uuid, item string) error {
				assert.Equal(t, "user-uuid-123", uuid)
				assert.Equal(t, "t-shirt", item)
				return nil
			},
		}
		controller.Storage = mockSt

		req := httptest.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
		w := httptest.NewRecorder()
		vars := map[string]string{
			"item": "t-shirt",
		}
		req = mux.SetURLVars(req, vars)

		controller.Buy(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("item not found -> 500", func(t *testing.T) {
		mockSt := &mockStorage{
			BuyFunc: func(ctx context.Context, uuid, item string) error {
				return storage.ErrItemNotFound
			},
		}
		controller.Storage = mockSt

		req := httptest.NewRequest(http.MethodGet, "/api/buy/unknown", nil)
		vars := map[string]string{"item": "unknown"}
		req = mux.SetURLVars(req, vars)

		w := httptest.NewRecorder()
		w.Header().Set("Authorization", "user-uuid-123")

		controller.Buy(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("not enough balance -> 500", func(t *testing.T) {
		mockSt := &mockStorage{
			BuyFunc: func(ctx context.Context, uuid, item string) error {
				return storage.ErrNotEnoughBalance
			},
		}
		controller.Storage = mockSt

		req := httptest.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
		vars := map[string]string{"item": "t-shirt"}
		req = mux.SetURLVars(req, vars)

		w := httptest.NewRecorder()
		w.Header().Set("Authorization", "user-uuid-123")

		controller.Buy(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("other error -> 500", func(t *testing.T) {
		mockSt := &mockStorage{
			BuyFunc: func(ctx context.Context, uuid, item string) error {
				return errors.New("some error")
			},
		}
		controller.Storage = mockSt

		req := httptest.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
		vars := map[string]string{"item": "t-shirt"}
		req = mux.SetURLVars(req, vars)

		w := httptest.NewRecorder()
		w.Header().Set("Authorization", "user-uuid-123")

		controller.Buy(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
