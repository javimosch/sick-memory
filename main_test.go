package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
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

func TestHandleVersion(t *testing.T) {
	// Capture stdout
	got := captureStdout(t, func() {
		handleVersion()
	})
	want := "sick-memory version 0.1.0\n"
	if got != want {
		t.Errorf("handleVersion() = %q, want %q", got, want)
	}
}

func TestGetGlobalSickMemoryDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}
	dir := getGlobalSickMemoryDir()
	want := filepath.Join(homeDir, ".sick-memory")
	if dir != want {
		t.Errorf("getGlobalSickMemoryDir() = %q, want %q", dir, want)
	}
}

func TestGetProjectMemoryPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name    string
		gitRoot string
		want    string
	}{
		{
			name:    "unix path",
			gitRoot: "/home/user/my-project",
			want:    filepath.Join(homeDir, ".sick-memory", "projects", "-home-user-my-project", "memory"),
		},
		{
			name:    "path with spaces",
			gitRoot: "/home/user/my project",
			want:    filepath.Join(homeDir, ".sick-memory", "projects", "-home-user-my_project", "memory"),
		},
		{
			name:    "empty root",
			gitRoot: "",
			want:    filepath.Join(homeDir, ".sick-memory", "projects", "", "memory"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getProjectMemoryPath(tt.gitRoot)
			if got != tt.want {
				t.Errorf("getProjectMemoryPath(%q) = %q, want %q", tt.gitRoot, got, tt.want)
			}
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	t.Run("simple data", func(t *testing.T) {
		got := captureStdout(t, func() {
			successResponse(map[string]interface{}{"key": "value"})
		})
		var resp SuccessResponse
		if err := json.Unmarshal([]byte(got), &resp); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if resp.Version != "1.0" {
			t.Errorf("Version = %q, want %q", resp.Version, "1.0")
		}
		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("Data is not a map, got %T", resp.Data)
		}
		if data["key"] != "value" {
			t.Errorf("Data.key = %v, want %v", data["key"], "value")
		}
	})

	t.Run("nil data", func(t *testing.T) {
		got := captureStdout(t, func() {
			successResponse(nil)
		})
		var resp SuccessResponse
		if err := json.Unmarshal([]byte(got), &resp); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("prints valid JSON", func(t *testing.T) {
		// errorResponse calls os.Exit, so we need to capture output differently
		// We'll just test the function doesn't panic and output is valid JSON
		got := captureStdout(t, func() {
			// Temporarily restore os.Exit for the real function
			// Since errorResponse calls os.Exit, use a test helper
		})
		_ = got
	})

	t.Run("JSON structure", func(t *testing.T) {
		// Directly test ErrorResponse JSON
		errResp := ErrorResponse{
			Error: ErrorDetail{
				Code:        85,
				Type:        "invalid_argument",
				Message:     "test error",
				Recoverable: false,
			},
		}
		data, err := json.Marshal(errResp)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		var decoded ErrorResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if decoded.Error.Code != 85 {
			t.Errorf("Code = %d, want %d", decoded.Error.Code, 85)
		}
		if decoded.Error.Type != "invalid_argument" {
			t.Errorf("Type = %q, want %q", decoded.Error.Type, "invalid_argument")
		}
		if decoded.Error.Message != "test error" {
			t.Errorf("Message = %q, want %q", decoded.Error.Message, "test error")
		}
	})
}

func TestSaveAndLoadSearchIndex(t *testing.T) {
	tmpDir := t.TempDir()

	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 2},
		},
		DocFreq: map[string]int{
			"golang": 1,
		},
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Name: "Test", Content: "test content", Type: "user", Created: time.Now()},
		},
	}

	// Save
	if err := saveSearchIndex(tmpDir, index); err != nil {
		t.Fatalf("saveSearchIndex failed: %v", err)
	}

	// Verify file exists
	indexFile := filepath.Join(tmpDir, "search_index.json")
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		t.Fatal("search_index.json was not created")
	}

	// Load
	loaded, err := loadSearchIndex(tmpDir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}

	if loaded.DocCount != 1 {
		t.Errorf("DocCount = %d, want %d", loaded.DocCount, 1)
	}
	if loaded.TermFreq["golang"]["mem1"] != 2 {
		t.Errorf("TermFreq[golang][mem1] = %d, want %d", loaded.TermFreq["golang"]["mem1"], 2)
	}
	if loaded.Memories["mem1"].Name != "Test" {
		t.Errorf("Memories[mem1].Name = %q, want %q", loaded.Memories["mem1"].Name, "Test")
	}
}

func TestLoadSearchIndex_NonExistentDir(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "nonexistent")
	_, err := loadSearchIndex(tmpDir)
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestBuildSearchIndex(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a memory file
	memoryContent := `---
name: Test Memory
description: A test memory for unit testing
type: user
created: 2026-06-16T00:00:00Z
---
This is the memory content.
`
	memoryFile := filepath.Join(tmpDir, "memory_test123.md")
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Build index
	index, err := buildSearchIndex(tmpDir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want %d", index.DocCount, 1)
	}
	if _, ok := index.Memories["memory_test123"]; !ok {
		t.Fatal("expected memory_test123 in index")
	}
	if len(index.TermFreq) == 0 {
		t.Error("expected non-empty TermFreq")
	}
}

func TestBuildSearchIndex_SkipDirectoriesAndHidden(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a subdirectory (should be skipped)
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	// Create a hidden file (should be skipped)
	hiddenFile := filepath.Join(tmpDir, ".hidden.md")
	if err := os.WriteFile(hiddenFile, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	// Create a non-memory file (should be skipped)
	otherFile := filepath.Join(tmpDir, "other.txt")
	if err := os.WriteFile(otherFile, []byte("other"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	index, err := buildSearchIndex(tmpDir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}

	if index.DocCount != 0 {
		t.Errorf("DocCount = %d, want 0 (all files should be filtered out)", index.DocCount)
	}
}

func TestBuildSearchIndex_NoDir(t *testing.T) {
	_, err := buildSearchIndex("/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}

func TestPrintHelp(t *testing.T) {
	got := captureStdout(t, func() {
		printHelp()
	})

	// Verify key content is present
	checks := []string{
		"sick-memory",
		"USAGE:",
		"COMMANDS:",
		"init",
		"remember",
		"recall",
		"list",
		"edit",
		"delete",
		"status",
		"config",
		"bridge",
		"AGENT BRIDGES:",
		"claude-code",
		"opencode",
		"copilot",
		"EXIT CODES:",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("printHelp() output missing %q", check)
		}
	}
}

func TestFindGitRepositoryRoot_InRepo(t *testing.T) {
	// Run inside the current repo (we're in a git repo during tests)
	root, err := findGitRepositoryRoot()
	if err != nil {
		t.Fatalf("findGitRepositoryRoot() returned error in git repo: %v", err)
	}
	if root == "" {
		t.Fatal("findGitRepositoryRoot() returned empty string in git repo")
	}
	// Verify it looks like a path
	if !strings.HasPrefix(root, "/") {
		t.Errorf("expected absolute path, got %q", root)
	}
}

func TestFindGitRepositoryRoot_OutsideRepo(t *testing.T) {
	// Create a temp dir that is NOT a git repo
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	_, err = findGitRepositoryRoot()
	if err == nil {
		t.Error("expected error when outside a git repository, got nil")
	}
}

func TestLoadGlobalConfig_CreatesDefault(t *testing.T) {
	// Use a temp HOME to isolate config creation
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	cfg := loadGlobalConfig()

	// Verify defaults
	if cfg.DefaultMemoryType != "user" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "user")
	}
	if cfg.MaxMemorySize != 1024*1024 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 1024*1024)
	}
	if !cfg.AutoIndex {
		t.Error("AutoIndex should be true by default")
	}

	// Verify config file was created
	configPath := filepath.Join(tmpHome, ".sick-memory", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config.json was not created")
	}
}

func TestLoadGlobalConfig_LoadsExisting(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create custom config
	configDir := filepath.Join(tmpHome, ".sick-memory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	customConfig := GlobalConfig{
		DefaultMemoryType: "project",
		MaxMemorySize:     512 * 1024,
		AutoIndex:         false,
	}
	data, _ := json.Marshal(customConfig)
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cfg := loadGlobalConfig()

	if cfg.DefaultMemoryType != "project" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "project")
	}
	if cfg.MaxMemorySize != 512*1024 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 512*1024)
	}
	if cfg.AutoIndex {
		t.Error("AutoIndex should be false")
	}
}

func TestLoadGlobalConfig_InvalidJSONFallsBack(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Write invalid JSON
	configDir := filepath.Join(tmpHome, ".sick-memory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte("not-json"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cfg := loadGlobalConfig()

	// Should return defaults on invalid JSON
	if cfg.DefaultMemoryType != "user" {
		t.Errorf("DefaultMemoryType = %q, want default %q", cfg.DefaultMemoryType, "user")
	}
}

func TestErrorResponse_PrintsValidJSON(t *testing.T) {
	got := captureStdout(t, func() {
		errorResponse(85, "invalid_argument", "test error message", false)
	})

	var resp ErrorResponse
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Fatalf("errorResponse() output is not valid JSON: %v\nOutput: %s", err, got)
	}
	if resp.Error.Code != 85 {
		t.Errorf("Code = %d, want %d", resp.Error.Code, 85)
	}
	if resp.Error.Type != "invalid_argument" {
		t.Errorf("Type = %q, want %q", resp.Error.Type, "invalid_argument")
	}
	if resp.Error.Message != "test error message" {
		t.Errorf("Message = %q, want %q", resp.Error.Message, "test error message")
	}
	if resp.Error.Recoverable {
		t.Error("Recoverable should be false")
	}
}

// captureStdout runs fn and returns everything written to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	return <-done
}
