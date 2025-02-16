package jwtToken_test

import (
	"avito/internal/utils/jwtToken"
	"testing"
)

func TestBuildTokenAndGetUserID(t *testing.T) {
	expectedUserID := "test-user-id"

	token, err := jwtToken.BuidToken(expectedUserID)
	if err != nil {
		t.Fatalf("Can't create token: %v", err)
	}

	actualUserID := jwtToken.GetUserID(token)
	if actualUserID != expectedUserID {
		t.Errorf("Expect UserID: %s, got: %s", expectedUserID, actualUserID)
	}
}

func TestGetUserID_InvalidToken(t *testing.T) {
	invalidToken := "some.invalid.token"

	userID := jwtToken.GetUserID(invalidToken)
	if userID != "" {
		t.Errorf("Expect ` `, got: %s", userID)
	}
}
