package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestSearchMemories_EmptyQuery(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		Memories: make(map[string]Memory),
		DocCount: 1,
	}
	idx.Memories["test1"] = Memory{
		ID:      "test1",
		Name:    "Test Memory",
		Content: "some content here",
	}
	idx.Memories["test2"] = Memory{
		ID:      "test2",
		Name:    "Test Memory 2",
		Content: "more content",
	}

	results := searchMemories(idx, "")
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(results))
	}

	results = searchMemories(idx, "   ")
	if len(results) != 0 {
		t.Errorf("expected 0 results for whitespace-only query, got %d", len(results))
	}
}

func TestAmCloudAdminAccess(t *testing.T) {
	token := os.Getenv("AM_WORKER_TOKEN")
	if token == "" {
		t.Skip("AM_WORKER_TOKEN not set, skipping am-cloud admin access test")
	}

	url := "https://automaintainer.intrane.fr/api/agent/memories?repo=71a96b&token=" + token
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("am-cloud admin API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("am-cloud admin API returned status %d, expected 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected JSON content-type, got %q", ct)
	}
}
