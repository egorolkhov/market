package sendCoin

import (
	"avito/internal/cache"
	"avito/internal/storage"
	"avito/internal/utils/jwtToken"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct {
	SendFunc              func(ctx context.Context, uuid, toUser string, amount int) error
	GetUuidByUsernameFunc func(ctx context.Context, username string) (string, error)
}

func (m *mockStorage) Send(ctx context.Context, uuid, toUser string, amount int) error {
	return m.SendFunc(ctx, uuid, toUser, amount)
}

func (m *mockStorage) GetUuidByUsername(ctx context.Context, username string) (string, error) {
	return m.GetUuidByUsernameFunc(ctx, username)
}

func TestSendController_SendCoin(t *testing.T) {
	lfu := cache.NewLFUCache(10)

	mockSt := &mockStorage{}
	controller := &SendController{
		Storage: mockSt,
		Lfu:     lfu,
	}

	t.Run("invalid JSON -> 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBufferString(`{not valid json`))
		w := httptest.NewRecorder()

		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "should return 400 for invalid JSON")
	})

	t.Run("missing fields -> 400", func(t *testing.T) {
		body := map[string]interface{}{
			"amount": 50,
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))
		w := httptest.NewRecorder()

		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("no JWT -> 500", func(t *testing.T) {
		body := map[string]interface{}{
			"toUser": "bob",
			"amount": 50,
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))
		w := httptest.NewRecorder()

		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("send success -> 200", func(t *testing.T) {
		mockSt.SendFunc = func(ctx context.Context, uuid, toUser string, amount int) error {
			assert.Equal(t, "uuid-123", uuid)
			assert.Equal(t, "bob", toUser)
			assert.Equal(t, 100, amount)
			return nil
		}

		mockSt.GetUuidByUsernameFunc = func(ctx context.Context, username string) (string, error) {
			assert.Equal(t, "bob", username)
			return "receiver-uuid-456", nil
		}

		body := map[string]interface{}{
			"toUser": "bob",
			"amount": 100,
		}
		b, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))

		token, _ := jwtToken.BuidToken("uuid-123")
		req.Header.Set("Authorization", token)

		w := httptest.NewRecorder()
		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("not enough balance -> 500", func(t *testing.T) {
		mockSt.SendFunc = func(ctx context.Context, uuid, toUser string, amount int) error {
			return storage.ErrNotEnoughBalance
		}
		mockSt.GetUuidByUsernameFunc = func(ctx context.Context, username string) (string, error) {
			return "receiver-uuid", nil
		}

		body := map[string]interface{}{
			"toUser": "bob",
			"amount": 9999,
		}
		b, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))
		token, _ := jwtToken.BuidToken("uuid-123")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("sending to yourself -> 500", func(t *testing.T) {
		mockSt.SendFunc = func(ctx context.Context, uuid, toUser string, amount int) error {
			return storage.ErrSendingToYourself
		}
		mockSt.GetUuidByUsernameFunc = func(ctx context.Context, username string) (string, error) {
			return "receiver-uuid", nil
		}

		body := map[string]interface{}{
			"toUser": "uuid-123",
			"amount": 10,
		}
		b, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))
		token, _ := jwtToken.BuidToken("uuid-123")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("user not found -> 500", func(t *testing.T) {
		mockSt.SendFunc = func(ctx context.Context, uuid, toUser string, amount int) error {
			return storage.ErrUserNotFound
		}
		mockSt.GetUuidByUsernameFunc = func(ctx context.Context, username string) (string, error) {
			return "", storage.ErrUserNotFound
		}

		body := map[string]interface{}{
			"toUser": "nonexistent",
			"amount": 1,
		}
		b, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))
		token, _ := jwtToken.BuidToken("user-uuid-123")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("other error -> 500", func(t *testing.T) {
		mockSt.SendFunc = func(ctx context.Context, uuid, toUser string, amount int) error {
			return errors.New("some error")
		}
		mockSt.GetUuidByUsernameFunc = func(ctx context.Context, username string) (string, error) {
			return "receiver-uuid", nil
		}

		body := map[string]interface{}{
			"toUser": "bob",
			"amount": 100,
		}
		b, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))
		token, _ := jwtToken.BuidToken("user-uuid-123")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("GetUuidByUsername returns error but still 200", func(t *testing.T) {
		mockSt.SendFunc = func(ctx context.Context, uuid, toUser string, amount int) error {
			return nil
		}
		mockSt.GetUuidByUsernameFunc = func(ctx context.Context, username string) (string, error) {
			return "", errors.New("some error")
		}
		lfu.Set("any-key", "any-value")

		body := map[string]interface{}{
			"toUser": "Bob",
			"amount": 10,
		}
		b, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBuffer(b))
		token, _ := jwtToken.BuidToken("uuid-123")
		req.Header.Set("Authorization", token)

		w := httptest.NewRecorder()
		controller.SendCoin(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_, found := lfu.Get("any-key")
		assert.False(t, found, "cache should be cleared")
	})
}
