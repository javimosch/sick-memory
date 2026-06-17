package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// cloudAPITestBase is the base URL for the am-cloud admin HTTP API.
// The WebSocket panel URL (AM_PANEL_URL) is separate; the admin API
// lives under an HTTPS endpoint on the same host.
const cloudAPITestBase = "https://automaintainer.intrane.fr/api"

// skipIfNoCloudEnv skips the test when the am-cloud environment
// variables required to reach the admin API are absent.
func skipIfNoCloudEnv(t *testing.T) (token, repo string) {
	t.Helper()
	token = os.Getenv("AM_WORKER_TOKEN")
	if token == "" {
		token = os.Getenv("AM_AGENT_TOKEN")
	}
	repo = os.Getenv("AM_REPO")
	if repo == "" {
		repo = "71a96b"
	}
	if token == "" {
		t.Skip("SKIP: neither AM_WORKER_TOKEN nor AM_AGENT_TOKEN is set")
	}
	return token, repo
}

func TestCloudAdminAPIReachable(t *testing.T) {
	token, repo := skipIfNoCloudEnv(t)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(cloudAPITestBase + "/agent/memories?repo=" + repo + "&token=" + token)
	if err != nil {
		t.Fatalf("am-cloud admin API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from am-cloud admin API, got %d", resp.StatusCode)
	}
}

func TestCloudAdminAPIAuthValid(t *testing.T) {
	token, repo := skipIfNoCloudEnv(t)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(cloudAPITestBase + "/agent/memories?repo=" + repo + "&token=" + token)
	if err != nil {
		t.Fatalf("am-cloud admin API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK with valid token, got %d", resp.StatusCode)
	}

	// Verify response is valid JSON
	var payload interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("expected valid JSON response from am-cloud admin API: %v", err)
	}
}

func TestCloudAdminAPIAuthRejectsBadToken(t *testing.T) {
	token, repo := skipIfNoCloudEnv(t)

	client := &http.Client{Timeout: 10 * time.Second}
	badToken := token + "_INVALID"
	resp, err := client.Get(cloudAPITestBase + "/agent/memories?repo=" + repo + "&token=" + badToken)
	if err != nil {
		t.Fatalf("am-cloud admin API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatal("expected auth failure with bad token, got 200 OK")
	}

	// Verify the error body is valid JSON (not an HTML error page)
	var payload interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("expected valid JSON error response: %v", err)
	}
}

func TestCloudAdminTokenFormat(t *testing.T) {
	token, _ := skipIfNoCloudEnv(t)

	if token == "" {
		t.Fatal("token should not be empty here")
	}
	if !strings.HasPrefix(token, "am_") {
		t.Errorf("am-cloud token should start with 'am_', got prefix: %s", token[:minInteger(5, len(token))])
	}
	if len(token) < 20 {
		t.Errorf("am-cloud token seems too short (%d chars, expected >= 20)", len(token))
	}
}

func TestCloudAdminPanelURLFormat(t *testing.T) {
	url := os.Getenv("AM_PANEL_URL")
	if url == "" {
		t.Skip("SKIP: AM_PANEL_URL not set")
	}
	if !strings.HasPrefix(url, "wss://") && !strings.HasPrefix(url, "https://") {
		t.Errorf("AM_PANEL_URL should start with wss:// or https://, got: %s", url)
	}
}

// minInteger returns the smaller of a and b.
func minInteger(a, b int) int {
	if a < b {
		return a
	}
	return b
}
