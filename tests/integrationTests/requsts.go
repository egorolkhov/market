package integrationTests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func doPost(t *testing.T, url string, body any, token string) (*http.Response, error) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Cookie", "auth_token="+token)
	}

	client := &http.Client{}
	return client.Do(req)
}

func doGet(t *testing.T, url string, token string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if token != "" {
		req.Header.Set("Cookie", "auth_token="+token)
	}
	client := &http.Client{}
	return client.Do(req)
}
