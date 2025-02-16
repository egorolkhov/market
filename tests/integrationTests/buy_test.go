package integrationTests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBuyItem(t *testing.T) {
	srv := setupTestServer(t)
	baseURL := srv.URL

	token := authUser(t, baseURL, "user", "password123")

	infoBefore := getInfo(t, baseURL, token)
	assert.Equal(t, 1000, infoBefore.Coins)

	buyURL := fmt.Sprintf("%s/api/buy/%s", baseURL, "t-shirt")
	resp, err := doGet(t, buyURL, token)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	infoAfter := getInfo(t, baseURL, token)

	assert.Equal(t, 1000-80, infoAfter.Coins)

	foundTShirt := false
	for _, inv := range infoAfter.Inventory {
		if inv.Type == "t-shirt" && inv.Quantity == 1 {
			foundTShirt = true
			break
		}
	}
	assert.True(t, foundTShirt)
}
