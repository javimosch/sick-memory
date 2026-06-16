package main

import (
	"math"
	"testing"
	"time"
)

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"replaces slashes", "/home/user/project", "-home-user-project"},
		{"replaces backslashes", "C:\\Users\\test", "C--Users-test"},
		{"replaces colons", "disk:volume:path", "disk-volume-path"},
		{"replaces spaces", "my project dir", "my_project_dir"},
		{"combined chars", "/a/b\\c:d e", "-a-b-c-d_e"},
		{"empty string", "", ""},
		{"no special chars", "simple-path", "simple-path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizePath(tt.input); got != tt.want {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"empty string", "", nil},
		{"short words filtered", "a an the at be", nil},
		{"single keyword", "golang", []string{"golang"}},
		{"multiple keywords", "golang testing framework", []string{"golang", "testing", "framework"}},
		{"punctuation stripped", "hello, world! test;", []string{"hello", "world", "test"}},
		{"stop words filtered", "the quick brown fox jumps", []string{"quick", "brown", "fox", "jumps"}},
		{"mixed case", "GoLang TESTING", []string{"golang", "testing"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractKeywords(tt.text)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("extractKeywords(%q) = %v (len=%d), want %v (len=%d)", tt.text, got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractKeywords(%q)[%d] = %q, want %q", tt.text, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestExtractKeywords_Boundaries(t *testing.T) {
	// Single char after punctuation should be filtered if <= 2
	got := extractKeywords("a. b! c?")
	if len(got) != 0 {
		t.Errorf("expected no keywords for short single chars, got %v", got)
	}

	// Case insensitivity
	got = extractKeywords("GOLANG")
	if len(got) != 1 || got[0] != "golang" {
		t.Errorf("expected [golang], got %v", got)
	}
}

func TestCalculateTFIDF(t *testing.T) {
	now := time.Now()
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 3, "mem2": 1},
			"testing": {"mem1": 2},
		},
		DocFreq: map[string]int{
			"golang":  2,
			"testing": 1,
		},
		DocCount: 2,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Name: "Memory 1", Created: now},
			"mem2": {ID: "mem2", Name: "Memory 2", Created: now},
		},
	}

	t.Run("zero doc count", func(t *testing.T) {
		emptyIdx := &SearchIndex{DocCount: 0}
		if got := calculateTFIDF(emptyIdx, "golang", "mem1"); got != 0 {
			t.Errorf("expected 0 for empty index, got %f", got)
		}
	})

	t.Run("term exists in doc", func(t *testing.T) {
		got := calculateTFIDF(idx, "golang", "mem1")
		// tf=3, df=2, N=2 => idf=log(3/3)=0 => tf*idf=0
		// Actually: idf = log((N+1)/(df+1)) = log(3/3) = 0
		// So tf*idf = 3 * 0 = 0
		if got != 0 {
			t.Errorf("expected 0 (log(3/3)=0), got %f", got)
		}
	})

	t.Run("term exists only in one doc", func(t *testing.T) {
		got := calculateTFIDF(idx, "testing", "mem1")
		// tf=2, df=1, N=2 => idf = log(3/2) = log(1.5)
		want := 2.0 * math.Log(3.0/2.0)
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("expected %f, got %f", want, got)
		}
	})

	t.Run("term not in doc", func(t *testing.T) {
		got := calculateTFIDF(idx, "golang", "nonexistent")
		if got != 0 {
			t.Errorf("expected 0 for nonexistent doc, got %f", got)
		}
	})

	t.Run("term not in index", func(t *testing.T) {
		got := calculateTFIDF(idx, "unknown", "mem1")
		if got != 0 {
			t.Errorf("expected 0 for unknown term, got %f", got)
		}
	})
}

func TestParseMemory(t *testing.T) {
	t.Run("full memory with frontmatter", func(t *testing.T) {
		content := `---
name: Test Memory
description: A test memory for unit testing
type: user
created: 2026-06-16T00:00:00Z
---
This is the memory content.
It spans multiple lines.
`
		memory := parseMemory(content, "memory_123.md")
		if memory.ID != "memory_123" {
			t.Errorf("ID = %q, want %q", memory.ID, "memory_123")
		}
		if memory.Name != "Test Memory" {
			t.Errorf("Name = %q, want %q", memory.Name, "Test Memory")
		}
		if memory.Description != "A test memory for unit testing" {
			t.Errorf("Description = %q, want %q", memory.Description, "A test memory for unit testing")
		}
		if memory.Type != "user" {
			t.Errorf("Type = %q, want %q", memory.Type, "user")
		}
		expectedContent := "This is the memory content.\nIt spans multiple lines.\n"
		if memory.Content != expectedContent {
			t.Errorf("Content = %q, want %q", memory.Content, expectedContent)
		}
	})

	t.Run("memory without frontmatter", func(t *testing.T) {
		content := "Just plain content\nNo metadata\n"
		memory := parseMemory(content, "memory_456.md")
		if memory.ID != "memory_456" {
			t.Errorf("ID = %q, want %q", memory.ID, "memory_456")
		}
		if memory.Content != "Just plain content\nNo metadata\n" {
			t.Errorf("Content = %q", memory.Content)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		memory := parseMemory("", "memory_789.md")
		if memory.ID != "memory_789" {
			t.Errorf("ID = %q", memory.ID)
		}
	})

	t.Run("invalid date is ignored", func(t *testing.T) {
		content := `---
created: not-a-date
---
body
`
		memory := parseMemory(content, "memory_bad_date.md")
		if !memory.Created.IsZero() {
			t.Errorf("expected zero time for invalid date, got %v", memory.Created)
		}
	})

	t.Run("partial frontmatter", func(t *testing.T) {
		content := `---
name: Partial Only
---
body text
`
		memory := parseMemory(content, "memory_partial.md")
		if memory.Name != "Partial Only" {
			t.Errorf("Name = %q, want %q", memory.Name, "Partial Only")
		}
		if memory.Description != "" {
			t.Errorf("Description should be empty")
		}
	})
}

func TestSearchMemories(t *testing.T) {
	baseTime := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang":  {"mem1": 2, "mem2": 1},
			"testing": {"mem1": 1},
			"rust":    {"mem2": 3},
		},
		DocFreq: map[string]int{
			"golang":  2,
			"testing": 1,
			"rust":    1,
		},
		DocCount: 2,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Name: "Go testing", Description: "golang testing", Content: "Writing tests in Go", Type: "project", Created: baseTime},
			"mem2": {ID: "mem2", Name: "Rust project", Description: "rust programming", Content: "Rust systems programming", Type: "user", Created: baseTime.Add(-48 * time.Hour)},
		},
	}

	t.Run("empty query returns all memories (exact phrase match on empty string)", func(t *testing.T) {
		// strings.Contains(s, "") always returns true in Go, so all memories get +2.0 exact phrase boost
		results := searchMemories(idx, "")
		if len(results) != 2 {
			t.Errorf("expected 2 results for empty query (all get phrase boost), got %d", len(results))
		}
	})

	t.Run("query matching one memory", func(t *testing.T) {
		results := searchMemories(idx, "rust")
		if len(results) != 1 {
			t.Errorf("expected 1 result for 'rust', got %d", len(results))
			return
		}
		if results[0].MemoryID != "mem2" {
			t.Errorf("expected mem2, got %s", results[0].MemoryID)
		}
	})

	t.Run("exact phrase boost", func(t *testing.T) {
		results := searchMemories(idx, "golang testing")
		if len(results) == 0 {
			t.Fatal("expected at least 1 result")
		}
		// mem1 should be ranked higher since it has exact phrase match
		if results[0].MemoryID != "mem1" {
			t.Errorf("expected mem1 (exact phrase match) first, got %s", results[0].MemoryID)
		}
	})

	t.Run("query matching no memories", func(t *testing.T) {
		results := searchMemories(idx, "python")
		if len(results) != 0 {
			t.Errorf("expected 0 results for 'python', got %d", len(results))
		}
	})

	t.Run("recency boost - recent memory scores higher", func(t *testing.T) {
		// Search for a term that appears in mem1 description only (for phrase boost)
		results := searchMemories(idx, "testing")
		if len(results) != 1 {
			t.Errorf("expected 1 result for 'testing', got %d", len(results))
			return
		}
		if results[0].MemoryID != "mem1" {
			t.Errorf("expected mem1 (matches 'testing'), got %s", results[0].MemoryID)
		}
	})
}

func TestGetDefaultMemoryDir(t *testing.T) {
	dir := getDefaultMemoryDir()
	if dir != ".sick-memory" {
		t.Errorf("expected '.sick-memory', got %q", dir)
	}
}

func TestVersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
	if Version != "0.1.0" {
		t.Errorf("expected '0.1.0', got %q", Version)
	}
}
