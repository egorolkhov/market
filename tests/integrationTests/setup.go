package integrationTests

import (
	"avito/internal/app"
	"avito/internal/app/routes"
	"avito/internal/app/services/auth"
	"avito/internal/cache"
	"avito/internal/config"
	"avito/internal/logger"
	"avito/internal/models"
	"avito/internal/storage"
	"avito/internal/utils"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	cfg := &config.Config{
		ServerAddress: ":0",
		PostgresConn:  "postgres://postgres:Qazxsw2200@postgres:5432/avito?sslmode=disable",
	}
	err := logger.InitLogger()
	assert.NoError(t, err, "Failed to init logger")

	lfu := cache.NewLFUCache(10)

	db := storage.NewStorage(storage.GetDatabaseDSN(*cfg))
	assert.NoError(t, err)
	err = utils.GooseDown(db.Tm.DB)
	err = utils.GooseUp(db.Tm.DB)
	assert.NoError(t, err)

	app := app.NewApp(db, lfu)

	h := routes.NewHandler(*app)
	srv := httptest.NewServer(h)

	t.Cleanup(func() {
		_ = db.Tm.DB.Close()
		srv.Close()
	})
	return srv
}

func authUser(t *testing.T, baseURL, username, password string) string {
	t.Helper()

	reqBody := auth.AuthRequest{
		Username: username,
		Password: password,
	}

	resp, err := doPost(t, baseURL+"/api/auth", reqBody, "")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var aResp auth.AuthResponse
	err = json.Unmarshal(body, &aResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, aResp.Token)

	return aResp.Token
}

func getInfo(t *testing.T, baseURL, token string) models.Info {
	t.Helper()

	resp, err := doGet(t, baseURL+"/api/info", token)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var info models.Info
	err = json.NewDecoder(resp.Body).Decode(&info)
	assert.NoError(t, err)
	return info
}
