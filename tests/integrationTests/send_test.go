package integrationTests

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func TestSendCoin(t *testing.T) {
	srv := setupTestServer(t)
	baseURL := srv.URL

	tokenAlice := authUser(t, baseURL, "user", "password123")
	tokenBob := authUser(t, baseURL, "user2", "password123")

	infoAliceBefore := getInfo(t, baseURL, tokenAlice)
	infoBobBefore := getInfo(t, baseURL, tokenBob)

	assert.Equal(t, 1000, infoAliceBefore.Coins)
	assert.Equal(t, 1000, infoBobBefore.Coins)

	sendBody := SendCoinRequest{
		ToUser: "user2",
		Amount: 200,
	}
	resp, err := doPost(t, baseURL+"/api/sendCoin", sendBody, tokenAlice)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	infoAliceAfter := getInfo(t, baseURL, tokenAlice)
	infoBobAfter := getInfo(t, baseURL, tokenBob)

	assert.Equal(t, 800, infoAliceAfter.Coins)
	assert.Equal(t, 1200, infoBobAfter.Coins)

	foundOutgoing := false
	for _, tx := range infoAliceAfter.CoinsHistory.Sent {
		if tx.ToUser == "user2" && tx.Amount == 200 {
			foundOutgoing = true
			break
		}
	}
	assert.True(t, foundOutgoing)

	foundIncoming := false
	for _, tx := range infoBobAfter.CoinsHistory.Received {
		if tx.FromUser == "user" && tx.Amount == 200 {
			foundIncoming = true
			break
		}
	}
	assert.True(t, foundIncoming)
}
