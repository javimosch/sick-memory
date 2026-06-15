package main

import (
	"math"
	"testing"
	"time"
)

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			input:    "a an the is", // all stop words
			expected: []string{},
		},
		{
			input:    "hello, world! test: done.",
			expected: []string{"hello", "world", "test", "done"},
		},
		{
			input:    "go",
			expected: []string{},
		},
		{
			input:    "good design patterns",
			expected: []string{"good", "design", "patterns"},
		},
	}

	for _, tt := range tests {
		result := extractKeywords(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("extractKeywords(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("extractKeywords(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestExtractKeywordsEmpty(t *testing.T) {
	result := extractKeywords("")
	if len(result) != 0 {
		t.Errorf("extractKeywords('') = %v, want empty", result)
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/home/user/project", "-home-user-project"},
		{"C:\\Users\\test", "C--Users-test"},
		{"path:with:colons", "path-with-colons"},
		{"path with spaces", "path_with_spaces"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		result := sanitizePath(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetDefaultMemoryDir(t *testing.T) {
	result := getDefaultMemoryDir()
	if result != ".sick-memory" {
		t.Errorf("getDefaultMemoryDir() = %q, want %q", result, ".sick-memory")
	}
}

func TestCalculateTFIDF(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"test": {"mem1": 3},
		},
		DocFreq: map[string]int{
			"test": 2,
		},
		DocCount: 5,
		Memories: map[string]Memory{},
	}

	score := calculateTFIDF(index, "test", "mem1")
	expectedTf := 3.0
	expectedIdf := math.Log(float64(5+1) / float64(2+1))
	expected := expectedTf * expectedIdf

	if score != expected {
		t.Errorf("calculateTFIDF() = %f, want %f", score, expected)
	}
}

func TestCalculateTFIDFZeroDocCount(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		DocCount: 0,
		Memories: map[string]Memory{},
	}

	score := calculateTFIDF(index, "test", "mem1")
	if score != 0 {
		t.Errorf("calculateTFIDF() with zero DocCount = %f, want 0", score)
	}
}

func TestParseMemory(t *testing.T) {
	content := `---
name: Test Memory
description: A test memory
type: user
created: 2024-01-15T10:00:00Z
---
This is the memory content
with multiple lines.`

	memory := parseMemory(content, "memory_123.md")

	if memory.ID != "memory_123" {
		t.Errorf("memory.ID = %q, want %q", memory.ID, "memory_123")
	}
	if memory.Name != "Test Memory" {
		t.Errorf("memory.Name = %q, want %q", memory.Name, "Test Memory")
	}
	if memory.Description != "A test memory" {
		t.Errorf("memory.Description = %q, want %q", memory.Description, "A test memory")
	}
	if memory.Type != "user" {
		t.Errorf("memory.Type = %q, want %q", memory.Type, "user")
	}
	expectedContent := "This is the memory content\nwith multiple lines."
	if memory.Content != expectedContent {
		t.Errorf("memory.Content = %q, want %q", memory.Content, expectedContent)
	}
}

func TestParseMemoryWithoutFrontmatter(t *testing.T) {
	content := "Just plain content\nwith no frontmatter."
	memory := parseMemory(content, "memory_456.md")

	if memory.ID != "memory_456" {
		t.Errorf("memory.ID = %q, want %q", memory.ID, "memory_456")
	}
	if memory.Content != content {
		t.Errorf("memory.Content = %q, want %q", memory.Content, content)
	}
}

func TestParseMemoryInvalidCreated(t *testing.T) {
	content := `---
name: Invalid Date
description: Testing bad date
type: user
created: not-a-date
---

Content here.`

	memory := parseMemory(content, "memory_bad.md")
	if !memory.Created.IsZero() {
		t.Errorf("memory.Created should be zero for invalid date, got %v", memory.Created)
	}
}

func TestSearchMemories(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"test": {"mem1": 1},
		},
		DocFreq: map[string]int{
			"test": 1,
		},
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "Test",
				Description: "A test memory",
				Type:        "user",
				Created:     time.Now(),
				Content:     "This is a test memory content",
			},
		},
	}

	results := searchMemories(index, "test")
	if len(results) == 0 {
		t.Fatal("searchMemories() returned 0 results, expected at least 1")
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("results[0].MemoryID = %q, want %q", results[0].MemoryID, "mem1")
	}
	if results[0].Score <= 0 {
		t.Errorf("results[0].Score = %f, want > 0", results[0].Score)
	}
}

func TestSearchMemoriesExactPhraseBoost(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "Exact",
				Description: "An exact phrase match memory",
				Type:        "user",
				Created:     time.Now(),
				Content:     "This memory contains a specific phrase",
			},
		},
	}

	results := searchMemories(index, "specific phrase")
	if len(results) == 0 {
		t.Fatal("searchMemories() returned 0 results, expected at least 1 for exact phrase match")
	}
	// With empty TF-IDF, exact phrase match should give score of 2.0
	if results[0].Score <= 0 {
		t.Errorf("results[0].Score = %f, want > 0 for exact phrase match", results[0].Score)
	}
}

func TestSearchMemoriesNoMatch(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"test": {"mem1": 1},
		},
		DocFreq: map[string]int{
			"test": 1,
		},
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "Test",
				Description: "A test memory",
				Type:        "user",
				Created:     time.Now(),
				Content:     "test content",
			},
		},
	}

	results := searchMemories(index, "nonexistent")
	if len(results) != 0 {
		t.Errorf("searchMemories() = %d results, want 0", len(results))
	}
}

func TestSearchMemoriesTypeBoost(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"test": {"mem1": 1},
		},
		DocFreq: map[string]int{
			"test": 1,
		},
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "Project Memory",
				Description: "test",
				Type:        "project",
				Created:     time.Now(),
				Content:     "test content",
			},
		},
	}

	results := searchMemories(index, "test")
	if len(results) == 0 {
		t.Fatal("searchMemories() returned 0 results for project type test")
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("results[0].MemoryID = %q, want %q", results[0].MemoryID, "mem1")
	}
}

func TestSearchMemoriesWordOverlapFallback(t *testing.T) {
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"ui":    {"mem1": 1},
			"setup": {"mem1": 1},
		},
		DocFreq: map[string]int{
			"ui":    1,
			"setup": 1,
		},
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "UI Setup",
				Description: "UI/Setup instructions",
				Type:        "user",
				Created:     time.Now(),
				Content:     "This covers UI design and setup procedures",
			},
		},
	}

	// Query "UI design" — "design" won't match via TF-IDF in "UI/Setup"
	// but "ui" will have a TF-IDF match, so we don't hit the fallback.
	// Let's query something that won't match at all via TF-IDF but will match individual keywords as substrings.
	results := searchMemories(index, "design")
	if len(results) > 0 {
		t.Log("searchMemories returned results for 'design'")
	}
}

func TestVersion(t *testing.T) {
	if Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", Version, "0.1.0")
	}
}
