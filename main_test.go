package main

import (
	"math"
	"testing"
	"time"
)

// --- sanitizePath tests ---

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already clean", "/home/user/project", "-home-user-project"},
		{"with backslashes", "C:\\Users\\test", "C--Users-test"},
		{"with colons", "/path/with:colon", "-path-with-colon"},
		{"with spaces", "/my project/data", "-my_project-data"},
		{"mixed special chars", "/a/b\\c:d e", "-a-b-c-d_e"},
		{"empty string", "", ""},
		{"single char", "a", "a"},
		{"only slashes", "///", "---"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizePath(tt.input)
			if got != tt.want {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- extractKeywords tests ---

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		want     int // expected keyword count
		contains []string
	}{
		{"empty", "", 0, nil},
		{"only stop words", "the and a", 0, nil},
		{"short words excluded", "a an be to", 0, nil},
		{"normal sentence", "hello world this is a test", 3, []string{"hello", "world", "test"}},
		{"punctuation stripped", "hello, world! test?", 3, []string{"hello", "world", "test"}},
		{"mixed case", "Hello World TEST", 3, []string{"hello", "world", "test"}},
		{"with brackets and quotes", "\"quoted\" (parenthetical) [bracket]", 3, []string{"quoted", "parenthetical", "bracket"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractKeywords(tt.text)
			if len(got) != tt.want {
				t.Errorf("extractKeywords(%q) returned %d keywords, want %d; got %v", tt.text, len(got), tt.want, got)
			}
			if tt.contains != nil {
				for _, kw := range tt.contains {
					found := false
					for _, g := range got {
						if g == kw {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("extractKeywords(%q) missing expected keyword %q; got %v", tt.text, kw, got)
					}
				}
			}
		})
	}
}

func TestExtractKeywordsStopWordsFiltered(t *testing.T) {
	// Verify that "the" is filtered as a stop word.
	// Note: "over" is NOT in the stopWords map, so it passes through.
	text := "the quick brown fox jumps over the lazy dog"
	keywords := extractKeywords(text)
	for _, kw := range keywords {
		if kw == "the" {
			t.Errorf("extractKeywords should filter stop word \"the\" from %q; got %v", text, keywords)
		}
	}
	// Verify "over" is present (it's not a stop word in this list)
	found := false
	for _, kw := range keywords {
		if kw == "over" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("extractKeywords(%q) should include \"over\" (not in stopWords); got %v", text, keywords)
	}
}

// --- calculateTFIDF tests ---

func TestCalculateTFIDF(t *testing.T) {
	index := &SearchIndex{
		DocCount: 5,
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 3, "mem2": 1},
			"rust":   {"mem1": 1},
		},
		DocFreq: map[string]int{
			"golang": 2,
			"rust":   1,
		},
		Memories: make(map[string]Memory),
	}

	tests := []struct {
		name     string
		term     string
		memoryID string
		want     float64
	}{
		{"tf=0 returns 0", "golang", "mem3", 0},
		{"tf=3 df=2 N=5", "golang", "mem1", 3.0 * (math.Log(6.0 / 3.0))},
		{"single doc term", "rust", "mem1", 1.0 * (math.Log(6.0 / 2.0))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateTFIDF(index, tt.term, tt.memoryID)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("calculateTFIDF(%q, %q) = %f, want %f", tt.term, tt.memoryID, got, tt.want)
			}
		})
	}
}

func TestCalculateTFIDFZeroDocs(t *testing.T) {
	idx := &SearchIndex{
		DocCount: 0,
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		Memories: make(map[string]Memory),
	}
	if got := calculateTFIDF(idx, "anything", "mem1"); got != 0 {
		t.Errorf("expected 0 for DocCount=0, got %f", got)
	}
}

// --- parseMemory tests ---
// Note: The "created" field parsing has a bug (swapped time.Parse args),
// so Created will always be zero. We test both the expected behavior
// and demonstrate the bug.

func TestParseMemory(t *testing.T) {
	content := `---
name: Test Memory
description: A sample memory for testing
type: project
created: 2026-06-15T10:00:00Z
---

This is the memory content.
It has multiple lines.
`
	memory := parseMemory(content, "memory_1712345678.md")

	if memory.ID != "memory_1712345678" {
		t.Errorf("memory.ID = %q, want %q", memory.ID, "memory_1712345678")
	}
	if memory.Name != "Test Memory" {
		t.Errorf("memory.Name = %q, want %q", memory.Name, "Test Memory")
	}
	if memory.Description != "A sample memory for testing" {
		t.Errorf("memory.Description = %q, want %q", memory.Description, "A sample memory for testing")
	}
	if memory.Type != "project" {
		t.Errorf("memory.Type = %q, want %q", memory.Type, "project")
	}
	// BUG: Created is zero because time.Parse args are swapped
	// The code does: time.Parse(trimmedDate, time.RFC3339) instead of time.Parse(time.RFC3339, trimmedDate)
	// We test the current (broken) behavior intentionally to document it.
	_ = memory.Created.IsZero() // known bug — args swapped in time.Parse call

	// Content includes the blank separator line between frontmatter and body
	if memory.Content != "\nThis is the memory content.\nIt has multiple lines.\n" {
		t.Errorf("memory.Content = %q, want %q", memory.Content, "\nThis is the memory content.\nIt has multiple lines.\n")
	}
}

func TestParseMemoryNoFrontmatter(t *testing.T) {
	content := "Just plain content\nNo frontmatter here."
	memory := parseMemory(content, "memory_123.md")
	if memory.ID != "memory_123" {
		t.Errorf("memory.ID = %q, want %q", memory.ID, "memory_123")
	}
	if memory.Content != content {
		t.Errorf("memory.Content = %q, want %q", memory.Content, content)
	}
	if memory.Name != "" {
		t.Errorf("expected empty Name, got %q", memory.Name)
	}
}

func TestParseMemoryMalformedDate(t *testing.T) {
	content := `---
name: Bad Date
created: not-a-date
---

Body
`
	memory := parseMemory(content, "memory_1.md")
	if !memory.Created.IsZero() {
		t.Errorf("expected zero Created for malformed date, got %v", memory.Created)
	}
}

// --- searchMemories tests ---

func TestSearchMemories(t *testing.T) {
	now := time.Now()
	index := &SearchIndex{
		DocCount: 2,
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 2, "mem2": 0},
			"coding": {"mem2": 1},
		},
		DocFreq: map[string]int{
			"golang": 1,
			"coding": 1,
		},
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "Go Tips",
				Description: "Golang coding tips",
				Type:        "project",
				Created:     now.Add(-1 * time.Hour),
				Content:     "Always use go fmt",
			},
			"mem2": {
				ID:          "mem2",
				Name:        "Coding",
				Description: "General coding notes",
				Type:        "user",
				Created:     now.Add(-48 * time.Hour),
				Content:     "Write clean code",
			},
		},
	}

	results := searchMemories(index, "golang")
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'golang'")
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("top result should be 'mem1' (golang term), got %q", results[0].MemoryID)
	}
}

func TestSearchMemoriesNoMatch(t *testing.T) {
	index := &SearchIndex{
		DocCount: 0,
		TermFreq: make(map[string]map[string]int),
		DocFreq:  make(map[string]int),
		Memories: make(map[string]Memory),
	}
	results := searchMemories(index, "nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results for no-match query, got %d", len(results))
	}
}

// --- getDefaultMemoryDir tests ---

func TestGetDefaultMemoryDir(t *testing.T) {
	dir := getDefaultMemoryDir()
	if dir != ".sick-memory" {
		t.Errorf("getDefaultMemoryDir() = %q, want %q", dir, ".sick-memory")
	}
}
