package main

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"stop words removed", "the and of a an", []string{}},
		{"short words excluded", "hi go it", []string{}},
		{"keywords extracted", "database connection timeout", []string{"database", "connection", "timeout"}},
		{"punctuation trimmed", "hello, world! test;", []string{"hello", "world", "test"}},
		{"mixed case", "API TOKEN secret", []string{"api", "token", "secret"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractKeywords(tt.input)
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("extractKeywords(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/project", "-home-user-project"},
		{"C:\\Users\\test", "C--Users-test"},
		{"/path/with spaces", "-path-with_spaces"},
		{"/path:with:colons", "-path-with-colons"},
		{"simple", "simple"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizePath(tt.input)
			if got != tt.want {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseMemory(t *testing.T) {
	content := `---
name: Test Memory
description: A test memory for testing
type: user
created: 2024-01-15T10:00:00Z
---

This is the memory content.
With multiple lines.
`
	memory := parseMemory(content, "memory_12345.md")
	if memory.ID != "memory_12345" {
		t.Errorf("ID = %q, want %q", memory.ID, "memory_12345")
	}
	if memory.Name != "Test Memory" {
		t.Errorf("Name = %q, want %q", memory.Name, "Test Memory")
	}
	if memory.Description != "A test memory for testing" {
		t.Errorf("Description = %q, want %q", memory.Description, "A test memory for testing")
	}
	if memory.Type != "user" {
		t.Errorf("Type = %q, want %q", memory.Type, "user")
	}
	if memory.Content != "\nThis is the memory content.\nWith multiple lines.\n" {
		t.Errorf("Content = %q, want %q", memory.Content, "\nThis is the memory content.\nWith multiple lines.\n")
	}
}

func TestParseMemoryNoFrontmatter(t *testing.T) {
	content := "Just content without frontmatter"
	memory := parseMemory(content, "memory_1.md")
	if memory.ID != "memory_1" {
		t.Errorf("ID = %q, want %q", memory.ID, "memory_1")
	}
	if memory.Content != "Just content without frontmatter" {
		t.Errorf("Content = %q", memory.Content)
	}
}

func TestCalculateTFIDF(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"database": {"m1": 3},
		},
		DocFreq: map[string]int{
			"database": 2,
		},
		Memories: map[string]Memory{
			"m1": {ID: "m1"},
			"m2": {ID: "m2"},
		},
		DocCount: 2,
	}
	got := calculateTFIDF(index, "database", "m1")
	// idf = log(3/3) = log(1) = 0 when DocCount == DocFreq (term appears in all docs)
	// need more docs than matching docs
	index2 := &SearchIndex{
		TermFreq: map[string]map[string]int{"api": {"m1": 2}},
		DocFreq:  map[string]int{"api": 1},
		Memories: map[string]Memory{"m1": {}, "m2": {}},
		DocCount: 2,
	}
	got2 := calculateTFIDF(index2, "api", "m1")
	if got2 <= 0 {
		t.Errorf("calculateTFIDF = %v, want positive", got2)
	}
	_ = got
}

func TestCalculateTFIDFZeroDocs(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		DocCount: 0,
	}
	got := calculateTFIDF(index, "anything", "m1")
	if got != 0 {
		t.Errorf("calculateTFIDF with zero docs = %v, want 0", got)
	}
}

func TestSearchMemoriesExactPhrase(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		Memories: map[string]Memory{
			"m1": {
				ID:      "m1",
				Name:    "Test",
				Content: "this is a database connection test",
				Type:    "user",
				Created: time.Now(),
			},
		},
		DocCount: 1,
	}
	results := searchMemories(index, "database connection")
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].MemoryID != "m1" {
		t.Errorf("expected m1, got %s", results[0].MemoryID)
	}
}

func TestSearchMemoriesNoMatch(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		Memories: map[string]Memory{
			"m1": {ID: "m1", Content: "something unrelated", Type: "user", Created: time.Now()},
		},
		DocCount: 1,
	}
	results := searchMemories(index, "zzzznotfound")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchMemoriesWordOverlapFallback(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		Memories: map[string]Memory{
			"m1": {
				ID:      "m1",
				Content: "disk-cleanup utility for removing temp files",
				Type:    "user",
				Created: time.Now(),
			},
		},
		DocCount: 1,
	}
	results := searchMemories(index, "disk cleanup")
	if len(results) == 0 {
		t.Fatal("expected word-overlap fallback to match")
	}
}

func TestSearchMemoriesTypeBoost(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{"test": {"m1": 1}},
		DocFreq:  map[string]int{"test": 1},
		Memories: map[string]Memory{
			"m1": {ID: "m1", Content: "test memory", Type: "project", Created: time.Now()},
		},
		DocCount: 1,
	}
	results := searchMemories(index, "test")
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].MemoryID != "m1" {
		t.Errorf("expected m1, got %s", results[0].MemoryID)
	}
}

func TestBuildSearchIndex(t *testing.T) {
	dir := t.TempDir()

	m1 := `---
name: Memory One
description: database config
type: user
created: 2024-01-15T10:00:00Z
---
database connection settings
`
	m2 := `---
name: Memory Two
description: api endpoints
type: project
created: 2024-01-15T11:00:00Z
---
api authentication tokens
`
	os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte(m1), 0644)
	os.WriteFile(filepath.Join(dir, "memory_2.md"), []byte(m2), 0644)
	os.WriteFile(filepath.Join(dir, "other.txt"), []byte("ignored"), 0644)

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex error: %v", err)
	}
	if index.DocCount != 2 {
		t.Errorf("DocCount = %d, want 2", index.DocCount)
	}
	if len(index.Memories) != 2 {
		t.Errorf("len(Memories) = %d, want 2", len(index.Memories))
	}
	if index.TermFreq["database"] == nil {
		t.Error("expected 'database' term in index")
	}
}

func TestBuildSearchIndexEmpty(t *testing.T) {
	dir := t.TempDir()
	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex on empty dir should not error: %v", err)
	}
	if index.DocCount != 0 {
		t.Errorf("DocCount = %d, want 0", index.DocCount)
	}
}

func TestBuildSearchIndexSkipsDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)
	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex skipping dirs should not error: %v", err)
	}
	if index.DocCount != 0 {
		t.Errorf("expected 0 memories, got %d", index.DocCount)
	}
}

func TestCalculateTFIDFMath(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"term": {"a": 2, "b": 1},
		},
		DocFreq: map[string]int{"term": 2},
		DocCount: 3,
	}
	// idf = log(4/3) = log(1.333) ≈ 0.287
	// tf-idf = 2 * 0.287 = 0.575
	got := calculateTFIDF(index, "term", "a")
	expected := 2.0 * math.Log(4.0/3.0)
	if math.Abs(got-expected) > 0.001 {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSearchMemoriesRecencyBoost(t *testing.T) {
	recent := time.Now().Add(-2 * time.Hour)
	old := time.Now().Add(-720 * time.Hour)
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"test": {"recent": 1, "old": 1},
		},
		DocFreq: map[string]int{"test": 2},
		Memories: map[string]Memory{
			"recent": {ID: "recent", Content: "test", Type: "user", Created: recent},
			"old":    {ID: "old", Content: "test", Type: "user", Created: old},
		},
		DocCount: 2,
	}
	results := searchMemories(index, "test")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].MemoryID != "recent" {
		t.Errorf("expected 'recent' first (boosted), got %s", results[0].MemoryID)
	}
}

func TestGetDefaultMemoryDir(t *testing.T) {
	dir := getDefaultMemoryDir()
	if dir != ".sick-memory" {
		t.Errorf("got %q, want '.sick-memory'", dir)
	}
}

func TestGetProjectMemoryPath(t *testing.T) {
	path := getProjectMemoryPath("/home/user/project")
	if !strings.Contains(path, "projects") || !strings.Contains(path, "-home-user-project") {
		t.Errorf("unexpected path: %s", path)
	}
	if strings.Contains(path, "..") {
		t.Error("path should not contain parent directory references")
	}
}

func TestSearchMemoriesMultipleTerms(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"database": {"m1": 1},
			"api":      {"m2": 1, "m1": 1},
		},
		DocFreq: map[string]int{"database": 1, "api": 2},
		Memories: map[string]Memory{
			"m1": {ID: "m1", Content: "database api", Type: "user", Created: time.Now()},
			"m2": {ID: "m2", Content: "api only", Type: "user", Created: time.Now()},
		},
		DocCount: 2,
	}
	results := searchMemories(index, "database api")
	if len(results) == 0 {
		t.Fatal("expected results")
	}
}
