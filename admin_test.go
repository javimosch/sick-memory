package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const adminAPIBase = "https://automaintainer.intrane.fr/api"

var httpClient = &http.Client{Timeout: 10 * time.Second}

type adminMemory struct {
	ID      string `json:"id"`
	RepoID  string `json:"repo_id"`
	UserID  int    `json:"user_id"`
	Tag     string `json:"tag"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

type adminError struct {
	Error string `json:"error"`
}

func skipIfNoToken(t *testing.T) string {
	t.Helper()
	token := os.Getenv("AM_WORKER_TOKEN")
	if token == "" {
		token = os.Getenv("AM_AGENT_TOKEN")
	}
	if token == "" {
		t.Skip("SKIP: AM_WORKER_TOKEN / AM_AGENT_TOKEN not set")
	}
	return token
}

func TestAdminAPIReachable(t *testing.T) {
	token := skipIfNoToken(t)

	resp, err := httpClient.Get(adminAPIBase + "/agent/memories?repo=71a96b&token=" + token)
	if err != nil {
		t.Fatalf("API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminAPIAuthWithValidToken(t *testing.T) {
	token := skipIfNoToken(t)

	resp, err := httpClient.Get(adminAPIBase + "/agent/memories?repo=71a96b&token=" + token)
	if err != nil {
		t.Fatalf("API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var memories []adminMemory
	if err := json.NewDecoder(resp.Body).Decode(&memories); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if len(memories) == 0 {
		t.Fatal("expected at least one memory, got empty array")
	}

	for _, m := range memories {
		if m.ID == "" {
			t.Error("memory has empty ID")
		}
		if m.Content == "" {
			t.Error("memory has empty content")
		}
	}
}

func TestAdminAPIAuthWithInvalidToken(t *testing.T) {
	token := skipIfNoToken(t)

	badToken := token + "_INVALID"
	resp, err := http.Get(adminAPIBase + "/agent/memories?repo=71a96b&token=" + badToken)
	if err != nil {
		t.Fatalf("API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatal("expected auth failure, got 200")
	}

	var errResp adminError
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("expected JSON error response: %v", err)
	}

	if errResp.Error == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestAdminAPIAuthWithMissingToken(t *testing.T) {
	resp, err := httpClient.Get(adminAPIBase + "/agent/memories?repo=71a96b")

	if err != nil {
		t.Fatalf("API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatal("expected auth failure with missing token, got 200")
	}

	var errResp adminError
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("expected JSON error response: %v", err)
	}

	if !strings.Contains(errResp.Error, "token") {
		t.Fatalf("error message should mention 'token': got %q", errResp.Error)
	}
}

func TestAdminAPIAuthWithMissingRepo(t *testing.T) {
	token := skipIfNoToken(t)

	resp, err := httpClient.Get(adminAPIBase + "/agent/memories?token=" + token)
	if err != nil {
		t.Fatalf("API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatal("expected error with missing repo, got 200")
	}

	var errResp adminError
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("expected JSON error response: %v", err)
	}

	if !strings.Contains(errResp.Error, "repo") {
		t.Fatalf("error message should mention 'repo': got %q", errResp.Error)
	}
}

func TestAdminAPIMemoryContentFormat(t *testing.T) {
	token := skipIfNoToken(t)

	resp, err := httpClient.Get(adminAPIBase + "/agent/memories?repo=71a96b&token=" + token)
	if err != nil {
		t.Fatalf("API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var memories []adminMemory
	if err := json.NewDecoder(resp.Body).Decode(&memories); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	expectedKeys := []string{"id", "repo_id", "user_id", "tag", "content", "source"}
	for _, m := range memories {
		raw, _ := json.Marshal(m)
		var rawMap map[string]interface{}
		json.Unmarshal(raw, &rawMap)

		for _, key := range expectedKeys {
			if _, exists := rawMap[key]; !exists {
				t.Errorf("memory missing field %q", key)
			}
		}
	}
}
