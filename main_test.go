package main

import (
	"math"
	"testing"
	"time"
)

// ─── extractKeywords ─────────────────────────────────────────────────────────

func TestExtractKeywords_RemovesStopWords(t *testing.T) {
	text := "the quick brown fox jumps over the lazy dog"
	keywords := extractKeywords(text)
	// "the" is a stop word, so should be removed
	for _, kw := range keywords {
		if kw == "the" {
			t.Errorf("expected 'the' (stop word) to be removed, got %q", kw)
		}
	}
	// Non-stop words should remain
	expected := []string{"quick", "brown", "fox", "jumps", "over", "lazy", "dog"}
	if len(keywords) != len(expected) {
		t.Errorf("expected %d keywords, got %d: %v", len(expected), len(keywords), keywords)
	}
	for i, kw := range expected {
		if i >= len(keywords) || keywords[i] != kw {
			t.Errorf("keyword[%d]: expected %q, got %q", i, kw, keywords[i])
		}
	}
}

func TestExtractKeywords_ShortWords(t *testing.T) {
	text := "a an is be to of it go my no"
	keywords := extractKeywords(text)
	// All are 2 chars or less, should all be removed
	if len(keywords) != 0 {
		t.Errorf("expected 0 keywords from short words, got %d: %v", len(keywords), keywords)
	}
}

func TestExtractKeywords_RemovesPunctuation(t *testing.T) {
	text := "hello, world! testing; code: bugs?"
	keywords := extractKeywords(text)
	expected := []string{"hello", "world", "testing", "code", "bugs"}
	if len(keywords) != len(expected) {
		t.Errorf("expected %d keywords, got %d", len(expected), len(keywords))
	}
}

func TestExtractKeywords_MixedCase(t *testing.T) {
	text := "The Quick Brown Fox"
	keywords := extractKeywords(text)
	if len(keywords) != 3 {
		t.Fatalf("expected 3 keywords, got %d: %v", len(keywords), keywords)
	}
	// Should all be lowercase
	for _, kw := range keywords {
		if kw != "quick" && kw != "brown" && kw != "fox" {
			t.Errorf("unexpected keyword %q (should be lowercase)", kw)
		}
	}
}

func TestExtractKeywords_EmptyInput(t *testing.T) {
	if kw := extractKeywords(""); len(kw) != 0 {
		t.Errorf("expected empty result, got %v", kw)
	}
}

// ─── calculateTFIDF ──────────────────────────────────────────────────────────

func TestCalculateTFIDF_NoDocuments(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		DocCount: 0,
		Memories: make(map[string]Memory),
	}
	score := calculateTFIDF(idx, "hello", "mem1")
	if score != 0 {
		t.Errorf("expected 0 for empty index, got %f", score)
	}
}

func TestCalculateTFIDF_TermNotFound(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{"hello": {"mem1": 2}},
		DocFreq:  map[string]int{"hello": 1},
		DocCount: 1,
		Memories: map[string]Memory{"mem1": {ID: "mem1"}},
	}
	score := calculateTFIDF(idx, "nonexistent", "mem1")
	if score != 0 {
		t.Errorf("expected 0 for missing term, got %f", score)
	}
}

func TestCalculateTFIDF_SingleDoc(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{"hello": {"mem1": 3}},
		DocFreq:  map[string]int{"hello": 1},
		DocCount: 1,
		Memories: map[string]Memory{"mem1": {ID: "mem1"}},
	}
	score := calculateTFIDF(idx, "hello", "mem1")
	// tf = 3, idf = log((1+1)/(1+1)) = log(1) = 0
	if score != 0 {
		t.Errorf("expected 0 (idf=0 when term in all docs), got %f", score)
	}
}

func TestCalculateTFIDF_MultiDoc(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{"hello": {"mem1": 2, "mem2": 1}},
		DocFreq:  map[string]int{"hello": 2},
		DocCount: 3,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1"},
			"mem2": {ID: "mem2"},
			"mem3": {ID: "mem3"},
		},
	}
	score := calculateTFIDF(idx, "hello", "mem1")
	// tf = 2, idf = log((3+1)/(2+1)) = log(4/3)
	expected := 2.0 * math.Log(4.0/3.0)
	if math.Abs(score-expected) > 0.0001 {
		t.Errorf("expected %f, got %f", expected, score)
	}
}

// ─── sanitizePath ────────────────────────────────────────────────────────────

func TestSanitizePath_ReplacesSlashes(t *testing.T) {
	result := sanitizePath("/home/user/project")
	if result != "-home-user-project" {
		t.Errorf("expected '-home-user-project', got %q", result)
	}
}

func TestSanitizePath_ReplacesBackslashes(t *testing.T) {
	result := sanitizePath("C:\\Users" + "\\" + "test")
	if result != "C--Users-test" {
		t.Errorf("expected 'C--Users-test', got %q", result)
	}
}

func TestSanitizePath_ReplacesColonsAndSpaces(t *testing.T) {
	result := sanitizePath("/my project:test")
	if result != "-my_project-test" {
		t.Errorf("expected '-my_project-test', got %q", result)
	}
}

func TestSanitizePath_EmptyInput(t *testing.T) {
	if s := sanitizePath(""); s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

// ─── parseMemory ─────────────────────────────────────────────────────────────

func TestParseMemory_Basic(t *testing.T) {
	content := `---
name: Test Memory
description: A test memory entry
type: project
created: 2026-06-17T00:00:00Z
---

This is the memory content.
It spans multiple lines.
`
	memory := parseMemory(content, "memory_12345.md")
	if memory.ID != "memory_12345" {
		t.Errorf("expected ID 'memory_12345', got %q", memory.ID)
	}
	if memory.Name != "Test Memory" {
		t.Errorf("expected Name 'Test Memory', got %q", memory.Name)
	}
	if memory.Description != "A test memory entry" {
		t.Errorf("expected Description 'A test memory entry', got %q", memory.Description)
	}
	if memory.Type != "project" {
		t.Errorf("expected Type 'project', got %q", memory.Type)
	}
	if !memory.Created.Equal(time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected 2026-06-17T00:00:00Z, got %v", memory.Created)
	}
	expectedContent := "\nThis is the memory content.\nIt spans multiple lines.\n"
	if memory.Content != expectedContent {
		t.Errorf("unexpected content:\n  got:  %q\n  want: %q", memory.Content, expectedContent)
	}
}

func TestParseMemory_NoFrontmatter(t *testing.T) {
	content := "Just some plain text content without frontmatter."
	memory := parseMemory(content, "memory_999.md")
	if memory.ID != "memory_999" {
		t.Errorf("expected ID 'memory_999', got %q", memory.ID)
	}
	if memory.Content != content {
		t.Errorf("expected content to be unchanged, got %q", memory.Content)
	}
}

func TestParseMemory_EmptyFrontmatter(t *testing.T) {
	content := `---
---

Content after empty frontmatter.`
	memory := parseMemory(content, "memory_001.md")
	if memory.Name != "" {
		t.Errorf("expected empty Name, got %q", memory.Name)
	}
	if memory.Content != "\nContent after empty frontmatter." {
		t.Errorf("unexpected content: %q", memory.Content)
	}
}

func TestParseMemory_InvalidTimestamp(t *testing.T) {
	content := `---
name: Bad Timestamp
created: not-a-timestamp
---

Content.`
	memory := parseMemory(content, "memory_bad.md")
	if !memory.Created.IsZero() {
		t.Errorf("expected zero time for invalid timestamp, got %v", memory.Created)
	}
}

// ─── searchMemories ──────────────────────────────────────────────────────────

func TestSearchMemories_EmptyIndex(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		DocCount: 0,
		Memories: make(map[string]Memory),
	}
	results := searchMemories(idx, "test query")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchMemories_ExactPhraseBoost(t *testing.T) {
	now := time.Now()
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"memory": {"mem1": 1},
			"system": {"mem1": 1},
		},
		DocFreq: map[string]int{"memory": 1, "system": 1},
		DocCount: 2,
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "Memory System",
				Description: "About the memory system",
				Content:     "This is about the memory system and how it works",
				Created:     now,
			},
			"mem2": {
				ID:          "mem2",
				Name:        "Other Note",
				Description: "Unrelated note",
				Content:     "Something completely different",
				Created:     now,
			},
		},
	}
	// Search for exact phrase that matches mem1's content
	results := searchMemories(idx, "memory system")
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("expected mem1 as top result, got %s", results[0].MemoryID)
	}
	// Exact phrase match should give a score > 2.0 (TF-IDF for 'memory' + 'system' + boost)
	if results[0].Score <= 2.0 {
		t.Errorf("expected score > 2.0 (exact phrase boost + TF-IDF), got %f", results[0].Score)
	}
}

func TestSearchMemories_ResultsSorted(t *testing.T) {
	now := time.Now()
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"test": {"high": 1},
		},
		DocFreq: map[string]int{"test": 1},
		DocCount: 3,
		Memories: map[string]Memory{
			"high": {
				ID:      "high",
				Name:    "High Score",
				Content: "This has the test keyword",
				Created: now,
			},
			"low": {
				ID:      "low",
				Name:    "Low Score",
				Content: "No match here at all",
				Created: now,
			},
		},
	}
	results := searchMemories(idx, "test")
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result, got %d", len(results))
	}
	if results[0].MemoryID != "high" {
		t.Errorf("expected 'high', got %q", results[0].MemoryID)
	}
}

func TestSearchMemories_RecencyBoost(t *testing.T) {
	recent := time.Now().Add(-1 * time.Hour)  // < 24h
	old := time.Now().Add(-48 * time.Hour)     // > 24h but < 168h
	ancient := time.Now().Add(-720 * time.Hour) // > 168h

	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"common": {"recent": 1, "old": 1, "ancient": 1},
		},
		DocFreq: map[string]int{"common": 3},
		DocCount: 3,
		Memories: map[string]Memory{
			"recent":  {ID: "recent", Content: "common keyword", Created: recent, Type: "user"},
			"old":     {ID: "old", Content: "common keyword", Created: old, Type: "user"},
			"ancient": {ID: "ancient", Content: "common keyword", Created: ancient, Type: "user"},
		},
	}
	results := searchMemories(idx, "common")
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// Recent should be first (highest recency boost: 1.2x)
	if results[0].MemoryID != "recent" {
		t.Errorf("expected 'recent' as top result (1.2x boost), got %q", results[0].MemoryID)
	}
	// Old should be second (1.1x boost)
	if results[1].MemoryID != "old" {
		t.Errorf("expected 'old' as second result (1.1x boost), got %q", results[1].MemoryID)
	}
}

func TestSearchMemories_TypeBoost(t *testing.T) {
	now := time.Now()
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"test": {"project": 1, "user": 1},
		},
		DocFreq: map[string]int{"test": 2},
		DocCount: 2,
		Memories: map[string]Memory{
			"project": {ID: "project", Content: "test project memory", Created: now, Type: "project"},
			"user":    {ID: "user", Content: "test user memory", Created: now, Type: "user"},
		},
	}
	results := searchMemories(idx, "test")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Project type has 1.15x boost, so should be first
	if results[0].MemoryID != "project" {
		t.Errorf("expected 'project' as top result (1.15x type boost), got %q", results[0].MemoryID)
	}
}

func TestSearchMemories_WordOverlapFallback(t *testing.T) {
	now := time.Now()
	idx := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {
				ID:      "mem1",
				Content: "UI/Design patterns for the app",
				Created: now,
				Type:    "user",
			},
		},
	}
	// "design" should match via word-overlap (substring match in content)
	results := searchMemories(idx, "design")
	if len(results) == 0 {
		t.Fatal("expected at least 1 result from word-overlap fallback")
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("expected 'mem1' matched via word-overlap, got %q", results[0].MemoryID)
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score from word-overlap, got %f", results[0].Score)
	}
}

func TestSearchMemories_NoMatch_ReturnsEmpty(t *testing.T) {
	now := time.Now()
	idx := &SearchIndex{
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Content: "some random text", Created: now, Type: "user"},
		},
	}
	results := searchMemories(idx, "zxqywv")
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", len(results))
	}
}

// ─── buildSearchIndex ────────────────────────────────────────────────────────
// Note: buildSearchIndex reads from disk, so tested indirectly via
// the TF-IDF/search functions above which test the indexing logic in isolation.

// ─── Config & Helpers ────────────────────────────────────────────────────────

func TestGetDefaultMemoryDir(t *testing.T) {
	dir := getDefaultMemoryDir()
	if dir != ".sick-memory" {
		t.Errorf("expected '.sick-memory', got %q", dir)
	}
}
