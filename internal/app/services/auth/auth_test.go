package auth

import (
	"avito/internal/models"
	"avito/internal/utils/jwtToken"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type mockStorage struct {
	getUserFunc    func(ctx context.Context, username string) (*models.User, error)
	createUserFunc func(ctx context.Context, user *models.User) error
}

func (m *mockStorage) GetUser(ctx context.Context, username string) (*models.User, error) {
	return m.getUserFunc(ctx, username)
}

func (m *mockStorage) CreateUser(ctx context.Context, user *models.User) error {
	return m.createUserFunc(ctx, user)
}

func TestAuthController_Auth(t *testing.T) {
	controller := &AuthController{
		Storage: &mockStorage{},
	}

	sendRequest := func(method, url string, body []byte) (*httptest.ResponseRecorder, *http.Request) {
		req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		controller.Auth(rr, req)
		return rr, req
	}

	t.Run("invalid JSON body", func(t *testing.T) {
		body := []byte(`{"username":"test","password":"12345"`)
		rr, _ := sendRequest(http.MethodPost, "/auth", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing username or password", func(t *testing.T) {
		body := []byte(`{"username":"","password":""}`)
		rr, _ := sendRequest(http.MethodPost, "/api/auth", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("user == nil -> create new user", func(t *testing.T) {
		mockStrg := &mockStorage{
			getUserFunc: func(ctx context.Context, username string) (*models.User, error) {
				return nil, nil
			},
			createUserFunc: func(ctx context.Context, user *models.User) error {
				assert.Equal(t, "newuser", user.Username)
				return nil
			},
		}
		controller.Storage = mockStrg

		body := []byte(`{"username":"  NewUser  ","password":"somepass"}`)
		rr, _ := sendRequest(http.MethodPost, "/api/auth", body)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Token)

		cookies := rr.Result().Cookies()
		assert.NotEmpty(t, cookies)
		assert.Equal(t, "auth_token", cookies[0].Name)
		assert.Equal(t, resp.Token, cookies[0].Value)
	})

	t.Run("user exists -> wrong password", func(t *testing.T) {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.DefaultCost)
		existingUser := &models.User{
			UUID:         uuid.New(),
			Username:     "user",
			PasswordHash: string(hashedPass),
			Balance:      1000,
			CreatedAt:    time.Now(),
		}
		mockStrg := &mockStorage{
			getUserFunc: func(ctx context.Context, username string) (*models.User, error) {
				return existingUser, nil
			},
		}
		controller.Storage = mockStrg

		body := []byte(`{"username":"user","password":"wrong"}`)
		rr, _ := sendRequest(http.MethodPost, "/api/auth", body)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("user exists -> correct password", func(t *testing.T) {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		existingUser := &models.User{
			UUID:         uuid.New(),
			Username:     "user",
			PasswordHash: string(hashedPass),
			Balance:      1000,
			CreatedAt:    time.Now(),
		}

		mockStrg := &mockStorage{
			getUserFunc: func(ctx context.Context, username string) (*models.User, error) {
				return existingUser, nil
			},
		}
		controller.Storage = mockStrg

		body := []byte(`{"username":"user","password":"password"}`)
		rr, _ := sendRequest(http.MethodPost, "/api/auth", body)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Token)

		userID := jwtToken.GetUserID(resp.Token)
		assert.Equal(t, existingUser.UUID.String(), userID)

		cookies := rr.Result().Cookies()
		assert.NotEmpty(t, cookies)
		assert.Equal(t, "auth_token", cookies[0].Name)
		assert.Equal(t, resp.Token, cookies[0].Value)
	})
}
