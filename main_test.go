package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- extractKeywords tests ---

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple keywords",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "stop words are filtered",
			input:    "the and of a in is it to with",
			expected: nil,
		},
		{
			name:     "mixed stop words and keywords",
			input:    "the quick brown fox jumps",
			expected: []string{"quick", "brown", "fox", "jumps"},
		},
		{
			name:     "punctuation stripped",
			input:    "hello, world! test? (parentheses) [brackets]",
			expected: []string{"hello", "world", "test", "parentheses", "brackets"},
		},
		{
			name:     "short words excluded (<=2 chars)",
			input:    "a an in on at to be by up we my go no of it he",
			expected: nil,
		},
		{
			name:     "mixed case",
			input:    "The QUICK Brown Fox",
			expected: []string{"quick", "brown", "fox"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only punctuation",
			input:    "... !!! ??? ,,,",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractKeywords(tt.input)
			if !stringSliceEqual(got, tt.expected) {
				t.Errorf("extractKeywords(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// --- sanitizePath tests ---

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/foo/bar", "-foo-bar"},
		{`C:\Users\test`, "C--Users-test"},
		{"path with spaces", "path_with_spaces"},
		{"colon:separated", "colon-separated"},
		{"normal-path", "normal-path"},
		{"", ""},
		{"already_clean", "already_clean"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizePath(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- calculateTFIDF tests ---

func TestCalculateTFIDF(t *testing.T) {
	// "golang" appears in 1 of 3 docs -> idf > 0
	// "api" appears in all 3 docs -> idf = 0
	// "rust" does not appear in any doc
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 3, "mem2": 1},
			"api":    {"mem1": 1, "mem2": 2, "mem3": 1},
			"rust":   {},
		},
		DocFreq: map[string]int{
			"golang": 1,
			"api":    3,
			"rust":   0,
		},
		DocCount: 3,
	}

	t.Run("existing term scores non-zero", func(t *testing.T) {
		// golang: tf=3, df=1, idf=log((3+1)/(1+1))=log(2)≈0.693 → score=3*0.693≈2.08
		score := calculateTFIDF(index, "golang", "mem1")
		if score <= 0 {
			t.Errorf("expected positive score for golang in mem1, got %f", score)
		}
	})

	t.Run("no documents returns zero", func(t *testing.T) {
		emptyIdx := &SearchIndex{DocCount: 0}
		score := calculateTFIDF(emptyIdx, "anything", "any")
		if score != 0 {
			t.Errorf("expected 0 for empty index, got %f", score)
		}
	})

	t.Run("higher tf gives higher score", func(t *testing.T) {
		score1 := calculateTFIDF(index, "golang", "mem1") // tf=3
		score2 := calculateTFIDF(index, "golang", "mem2") // tf=1
		if score1 <= score2 {
			t.Errorf("expected mem1 (tf=3) > mem2 (tf=1), got %f vs %f", score1, score2)
		}
	})

	t.Run("term in all docs scores zero", func(t *testing.T) {
		// "api" has df=3 (all docs), so idf=log((3+1)/(3+1))=log(1)=0
		score := calculateTFIDF(index, "api", "mem1")
		if score != 0 {
			t.Errorf("expected 0 for term in all docs, got %f", score)
		}
	})

	t.Run("term with no frequency scores zero", func(t *testing.T) {
		score := calculateTFIDF(index, "rust", "mem1")
		if score != 0 {
			t.Errorf("expected 0 for term with no frequency, got %f", score)
		}
	})
}

// --- parseMemory tests ---

func TestParseMemory(t *testing.T) {
	t.Run("full frontmatter with content", func(t *testing.T) {
		content := `---
name: Test Memory
description: A test memory for unit testing
type: user
created: 2026-06-15T10:00:00Z
---
This is the memory content.
It spans multiple lines.

Second paragraph.`
		memory := parseMemory(content, "memory_12345.md")
		if memory.ID != "memory_12345" {
			t.Errorf("expected ID memory_12345, got %s", memory.ID)
		}
		if memory.Name != "Test Memory" {
			t.Errorf("expected Name 'Test Memory', got %q", memory.Name)
		}
		if memory.Description != "A test memory for unit testing" {
			t.Errorf("expected Description 'A test memory...', got %q", memory.Description)
		}
		if memory.Type != "user" {
			t.Errorf("expected Type 'user', got %q", memory.Type)
		}
		if memory.Content != "This is the memory content.\nIt spans multiple lines.\n\nSecond paragraph." {
			t.Errorf("unexpected content: %q", memory.Content)
		}
	})

	t.Run("missing frontmatter fields", func(t *testing.T) {
		content := `---
name: Minimal
---
Just content.`
		memory := parseMemory(content, "memory_678.md")
		if memory.ID != "memory_678" {
			t.Errorf("expected ID memory_678, got %s", memory.ID)
		}
		if memory.Name != "Minimal" {
			t.Errorf("expected Name 'Minimal', got %q", memory.Name)
		}
		if memory.Description != "" {
			t.Errorf("expected empty Description, got %q", memory.Description)
		}
		if memory.Type != "" {
			t.Errorf("expected empty Type, got %q", memory.Type)
		}
	})

	t.Run("no frontmatter", func(t *testing.T) {
		content := "Just raw content\nNo frontmatter here."
		memory := parseMemory(content, "memory_999.md")
		if memory.ID != "memory_999" {
			t.Errorf("expected ID memory_999, got %s", memory.ID)
		}
		if memory.Content != content {
			t.Errorf("expected content preserved, got %q", memory.Content)
		}
	})

	t.Run("empty created field", func(t *testing.T) {
		content := `---
name: No Date
created: 
---
Content`
		memory := parseMemory(content, "memory_111.md")
		// created field parsing: when value is empty string (after trimming),
		// time.Parse will fail and Created should be zero-value
		if !memory.Created.IsZero() {
			t.Errorf("expected zero Created time for empty field, got %v", memory.Created)
		}
	})
}

// --- searchMemories tests ---

func TestSearchMemories(t *testing.T) {
	// Create a fixed time for test stability
	fixedTime := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 2},
			"api":    {"mem1": 1, "mem2": 1},
			"design": {"mem2": 3},
		},
		DocFreq: map[string]int{
			"golang": 1,
			"api":    2,
			"design": 1,
		},
		DocCount: 2,
		Memories: map[string]Memory{
			"mem1": {
				ID:          "mem1",
				Name:        "Go Project",
				Description: "A golang api project",
				Type:        "project",
				Created:     fixedTime,
				Content:     "This is a golang API project with multiple endpoints.",
			},
			"mem2": {
				ID:          "mem2",
				Name:        "Design Docs",
				Description: "API design documentation",
				Type:        "user",
				Created:     fixedTime.Add(-48 * time.Hour), // 2 days old
				Content:     "Design patterns for the API layer.",
			},
		},
	}

	t.Run("matching query returns results", func(t *testing.T) {
		results := searchMemories(index, "golang api")
		if len(results) == 0 {
			t.Fatal("expected at least one result for 'golang api'")
		}
		// mem1 should rank highest (golang + api match + project boost + newer)
		if results[0].MemoryID != "mem1" {
			t.Errorf("expected mem1 as top result, got %s", results[0].MemoryID)
		}
	})

	t.Run("non-matching query returns empty", func(t *testing.T) {
		results := searchMemories(index, "nonexistent term xyz")
		if len(results) != 0 {
			t.Errorf("expected no results, got %d", len(results))
		}
	})

	t.Run("empty query matches all via substring containment", func(t *testing.T) {
		// strings.Contains(anything, "") returns true, so every memory gets
		// the exact phrase boost (+2.0). This is a known behavior — the
		// handleRecall function guards against empty query before calling
		// searchMemories.
		results := searchMemories(index, "")
		if len(results) != len(index.Memories) {
			t.Errorf("expected all memories returned for empty query, got %d", len(results))
		}
	})

	t.Run("exact phrase boost", func(t *testing.T) {
		results := searchMemories(index, "golang API project")
		if len(results) == 0 {
			t.Fatal("expected results for existing phrase")
		}
		// mem1 contains "golang API project" implicitly via "golang api project" in content
		// Actually: "golang" + "api" are separate words; exact match checks full string
		// "golang API project" appears? "This is a golang API project" - yes with casing diff
		// lowerContent="this is a golang api project with multiple endpoints."
		// lowerQuery="golang api project" -> contains? "golang api project" is in lowerContent as "golang api project"
		// Actually: "golang API project" lowered = "golang api project"
		// "This is a golang API project" lowered = "this is a golang api project with multiple endpoints."
		// Yes, substring match! So score includes +2.0 boost.
		hasBoost := false
		for _, r := range results {
			if r.MemoryID == "mem1" && r.Score > 2.0 {
				hasBoost = true
				break
			}
		}
		if !hasBoost {
			t.Errorf("expected mem1 to have exact phrase boost (score > 2.0)")
		}
	})
}

// --- loadGlobalConfig tests ---

func TestLoadGlobalConfig(t *testing.T) {
	t.Run("returns defaults when no config file", func(t *testing.T) {
		// Temporarily set HOME to a temp dir to avoid interference
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		config := loadGlobalConfig()
		if config.DefaultMemoryType != "user" {
			t.Errorf("expected default_memory_type 'user', got %q", config.DefaultMemoryType)
		}
		if config.MaxMemorySize != 1024*1024 {
			t.Errorf("expected MaxMemorySize 1048576, got %d", config.MaxMemorySize)
		}
		if !config.AutoIndex {
			t.Errorf("expected AutoIndex true, got false")
		}
	})

	t.Run("loads existing config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		// Create config directory and file
		configDir := filepath.Join(tmpDir, ".sick-memory")
		os.MkdirAll(configDir, 0755)
		configContent := `{
			"default_memory_type": "project",
			"max_memory_size": 512,
			"auto_index": false
		}`
		os.WriteFile(filepath.Join(configDir, "config.json"), []byte(configContent), 0644)

		config := loadGlobalConfig()
		if config.DefaultMemoryType != "project" {
			t.Errorf("expected default_memory_type 'project', got %q", config.DefaultMemoryType)
		}
		if config.MaxMemorySize != 512 {
			t.Errorf("expected MaxMemorySize 512, got %d", config.MaxMemorySize)
		}
		if config.AutoIndex {
			t.Errorf("expected AutoIndex false, got true")
		}
	})
}

// --- handleVersion test ---

func TestHandleVersion(t *testing.T) {
	// capture stdout
	rescue := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handleVersion()

	w.Close()
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	os.Stdout = rescue

	output := string(buf[:n])
	if output != "sick-memory version 0.1.0\n" {
		t.Errorf("unexpected version output: %q", output)
	}
}

// --- buildSearchIndex tests ---

func TestBuildSearchIndex(t *testing.T) {
	t.Run("empty directory returns empty index", func(t *testing.T) {
		tmpDir := t.TempDir()
		index, err := buildSearchIndex(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if index.DocCount != 0 {
			t.Errorf("expected DocCount 0, got %d", index.DocCount)
		}
	})

	t.Run("skips non-memory files", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create some non-memory files
		os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("readme"), 0644)
		os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644)

		index, err := buildSearchIndex(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if index.DocCount != 0 {
			t.Errorf("expected DocCount 0 for non-memory files, got %d", index.DocCount)
		}
	})

	t.Run("parses valid memory files", func(t *testing.T) {
		tmpDir := t.TempDir()
		memoryContent := `---
name: Test
description: A test memory for searching
type: project
created: 2026-06-15T10:00:00Z
---
Test content here.`
		os.WriteFile(filepath.Join(tmpDir, "memory_001.md"), []byte(memoryContent), 0644)

		index, err := buildSearchIndex(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if index.DocCount != 1 {
			t.Errorf("expected DocCount 1, got %d", index.DocCount)
		}
		if _, ok := index.Memories["memory_001"]; !ok {
			t.Errorf("expected memory_001 in index")
		}
		// Keywords should be extracted from description + content
		// "A test memory for searching" -> keywords: [test, memory, searching]
		// "Test content here" -> keywords: [test, content, here]
		// Combined unique: test, memory, searching, content, here
		if len(index.TermFreq) < 3 {
			t.Errorf("expected at least 3 terms in index, got %d", len(index.TermFreq))
		}
	})

	t.Run("skips unparseable files gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Write an invalid memory file that won't parse
		os.WriteFile(filepath.Join(tmpDir, "memory_bad.md"), []byte("no frontmatter at all"), 0644)
		// And a valid one
		validContent := `---
name: Valid
description: valid memory
type: user
---
Valid content`
		os.WriteFile(filepath.Join(tmpDir, "memory_002.md"), []byte(validContent), 0644)

		index, err := buildSearchIndex(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// memory_bad.md has no frontmatter (no ---), so parseMemory creates memory with ID="memory_bad" and no name/description
		// It should still be indexed since ID is not empty (it's derived from filename)
		if index.DocCount < 1 {
			t.Errorf("expected at least 1 memory indexed, got %d", index.DocCount)
		}
	})
}

// Helper for slice comparison
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
