package main

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBuildSearchIndexEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 0 {
		t.Errorf("DocCount = %d, want 0", index.DocCount)
	}
	if len(index.Memories) != 0 {
		t.Errorf("Memories count = %d, want 0", len(index.Memories))
	}
	if len(index.TermFreq) != 0 {
		t.Errorf("TermFreq count = %d, want 0", len(index.TermFreq))
	}
	if len(index.DocFreq) != 0 {
		t.Errorf("DocFreq count = %d, want 0", len(index.DocFreq))
	}
}

func TestBuildSearchIndex(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	writeMemoryFile(t, dir, "memory_2.md", `---
name: Memory Two
description: rust project
type: user
created: 2026-07-11T10:00:00Z
---
rust programming
`)

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 2 {
		t.Errorf("DocCount = %d, want 2", index.DocCount)
	}

	wantTermFreq := map[string]map[string]int{
		"golang":      {"memory_1": 2},
		"testing":     {"memory_1": 1},
		"tests":       {"memory_1": 1},
		"rust":        {"memory_2": 2},
		"project":     {"memory_2": 1},
		"programming": {"memory_2": 1},
	}
	if !reflect.DeepEqual(index.TermFreq, wantTermFreq) {
		t.Errorf("TermFreq = %v, want %v", index.TermFreq, wantTermFreq)
	}

	wantDocFreq := map[string]int{
		"golang":      1,
		"testing":     1,
		"tests":       1,
		"rust":        1,
		"project":     1,
		"programming": 1,
	}
	if !reflect.DeepEqual(index.DocFreq, wantDocFreq) {
		t.Errorf("DocFreq = %v, want %v", index.DocFreq, wantDocFreq)
	}

	if len(index.Memories) != 2 {
		t.Errorf("Memories count = %d, want 2", len(index.Memories))
	}
	for _, id := range []string{"memory_1", "memory_2"} {
		if _, ok := index.Memories[id]; !ok {
			t.Errorf("missing memory %s", id)
		}
	}
}

func TestBuildSearchIndexHandlesMemoryWithoutFrontmatter(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", "plain golang content without frontmatter")

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}
	if index.Memories["memory_1"].Content != "plain golang content without frontmatter" {
		t.Errorf("Content = %q, want %q", index.Memories["memory_1"].Content, "plain golang content without frontmatter")
	}
	if index.TermFreq["golang"]["memory_1"] != 1 {
		t.Errorf("TermFreq[golang][memory_1] = %d, want 1", index.TermFreq["golang"]["memory_1"])
	}
}

func TestBuildSearchIndexIgnoresNonMemoryFiles(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	for _, name := range []string{"README.md", "notes.txt", "memory_2.txt", ".hidden.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("should be ignored"), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}
	if _, ok := index.Memories["README"]; ok {
		t.Errorf("did not expect README.md to be indexed")
	}
}

func TestBuildSearchIndexIgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	subDir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "memory_2.md"), []byte("should be ignored"), 0644); err != nil {
		t.Fatalf("failed to write nested memory: %v", err)
	}

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}
	if _, ok := index.Memories["memory_2"]; ok {
		t.Errorf("did not expect nested memory_2.md to be indexed")
	}
}

func TestLoadSearchIndexUsesCacheFile(t *testing.T) {
	dir := t.TempDir()

	cache := `{
  "TermFreq": {"golang": {"memory_1": 2}},
  "DocFreq": {"golang": 1},
  "DocCount": 1,
  "Memories": {
    "memory_1": {
      "id": "memory_1",
      "name": "Cached Memory",
      "description": "cached",
      "type": "project",
      "created": "2026-07-11T12:00:00Z",
      "content": "cached content"
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "search_index.json"), []byte(cache), 0644); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	index, err := loadSearchIndex(dir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 from cache, got %v", index.Memories)
	}
	if index.Memories["memory_1"].Name != "Cached Memory" {
		t.Errorf("Name = %q, want %q", index.Memories["memory_1"].Name, "Cached Memory")
	}
}

func TestSaveAndLoadSearchIndex(t *testing.T) {
	dir := t.TempDir()

	created := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"memory_1": 2, "memory_2": 1},
			"rust":   {"memory_2": 3},
		},
		DocFreq: map[string]int{
			"golang": 2,
			"rust":   1,
		},
		DocCount: 2,
		Memories: map[string]Memory{
			"memory_1": {
				ID:          "memory_1",
				Name:        "Memory One",
				Description: "golang testing",
				Type:        "project",
				Created:     created,
				Content:     "Write tests in golang",
			},
			"memory_2": {
				ID:          "memory_2",
				Name:        "Memory Two",
				Description: "rust project",
				Type:        "user",
				Created:     created,
				Content:     "rust programming",
			},
		},
	}

	if err := saveSearchIndex(dir, index); err != nil {
		t.Fatalf("saveSearchIndex failed: %v", err)
	}

	loaded, err := loadSearchIndex(dir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected loaded index, got nil")
	}

	if loaded.DocCount != index.DocCount {
		t.Errorf("DocCount = %d, want %d", loaded.DocCount, index.DocCount)
	}

	if !reflect.DeepEqual(loaded.TermFreq, index.TermFreq) {
		t.Errorf("TermFreq = %v, want %v", loaded.TermFreq, index.TermFreq)
	}

	if !reflect.DeepEqual(loaded.DocFreq, index.DocFreq) {
		t.Errorf("DocFreq = %v, want %v", loaded.DocFreq, index.DocFreq)
	}

	if len(loaded.Memories) != len(index.Memories) {
		t.Fatalf("Memories count = %d, want %d", len(loaded.Memories), len(index.Memories))
	}
	for id, want := range index.Memories {
		got, ok := loaded.Memories[id]
		if !ok {
			t.Errorf("missing memory %s", id)
			continue
		}
		if got.ID != want.ID || got.Name != want.Name || got.Description != want.Description || got.Type != want.Type || got.Content != want.Content {
			t.Errorf("memory %s = %+v, want %+v", id, got, want)
		}
		if !got.Created.Equal(want.Created) {
			t.Errorf("memory %s Created = %v, want %v", id, got.Created, want.Created)
		}
	}
}

func TestBuildSearchIndexNonExistentDirectory(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist")

	index, err := buildSearchIndex(nonExistentPath)
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
	if index != nil {
		t.Fatalf("expected nil index on error, got %v", index)
	}
}

func TestLoadSearchIndexPropagatesBuildError(t *testing.T) {
	dir := t.TempDir()

	// Create a file where loadSearchIndex expects a directory.
	filePath := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	index, err := loadSearchIndex(filePath)
	if err == nil {
		t.Fatal("expected error when buildSearchIndex fails, got nil")
	}
	if index != nil {
		t.Fatalf("expected nil index on error, got %v", index)
	}
}

func TestLoadSearchIndexNonExistentDirectory(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist")

	index, err := loadSearchIndex(nonExistentPath)
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
	if index != nil {
		t.Fatalf("expected nil index on error, got %v", index)
	}
}

func TestLoadSearchIndexFallsBackOnCorruptCache(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "search_index.json"), []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write corrupt cache: %v", err)
	}

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	index, err := loadSearchIndex(dir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}
}

func TestLoadSearchIndexBuildsWhenCacheMissing(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	index, err := loadSearchIndex(dir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}

	// A cache file should be written for future loads.
	if _, err := os.Stat(filepath.Join(dir, "search_index.json")); err != nil {
		t.Errorf("expected cache file to be written: %v", err)
	}
}

func TestSaveSearchIndexMissingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")

	index := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		DocCount: 0,
		Memories: map[string]Memory{},
	}

	if err := saveSearchIndex(dir, index); err == nil {
		t.Fatal("saveSearchIndex expected error for missing directory, got nil")
	}
}

func TestBuildSearchIndexMemoryWithOnlyStopWords(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", "the a an of at")

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}
	if len(index.TermFreq) != 0 {
		t.Errorf("TermFreq count = %d, want 0", len(index.TermFreq))
	}
}

func TestBuildSearchIndexSkipsUnreadableMemoryFile(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", "golang content")

	broken := filepath.Join(dir, "memory_2.md")
	if err := os.Symlink("/nonexistent/path/broken", broken); err != nil {
		t.Skip("symlinks not supported in this test environment:", err)
	}

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}
	if _, ok := index.Memories["memory_2"]; ok {
		t.Errorf("did not expect broken memory_2.md to be indexed")
	}
}

func TestBuildSearchIndexSkipsEmptyMemoryID(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", "golang content")

	if err := os.WriteFile(filepath.Join(dir, ".md"), []byte("empty filename"), 0644); err != nil {
		t.Fatalf("failed to write .md file: %v", err)
	}

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed, got %v", index.Memories)
	}
	if _, ok := index.Memories[""]; ok {
		t.Errorf("did not expect empty memory ID to be indexed")
	}
}

func TestBuildSearchIndexPopulatesCreated(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	want := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	got := index.Memories["memory_1"].Created
	if !got.Equal(want) {
		t.Errorf("Created = %v, want %v", got, want)
	}
}

func writeMemoryFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

func TestSearchMemoriesExactPhrase(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", "golang testing project")
	writeMemoryFile(t, dir, "memory_2.md", "golang project")

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	results := searchMemories(index, "golang testing")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].MemoryID != "memory_1" {
		t.Errorf("expected exact-phrase memory memory_1 to rank first, got %s", results[0].MemoryID)
	}
	if results[0].Score <= results[1].Score {
		t.Errorf("expected exact-phrase score %v to be higher than non-phrase score %v", results[0].Score, results[1].Score)
	}
}

func TestSearchMemoriesWordOverlapFallback(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang-testing": {"mem1": 1},
		},
		DocFreq: map[string]int{
			"golang-testing": 1,
		},
		DocCount: 1,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Name: "Memory", Description: "", Content: "golang-testing", Type: "user", Created: time.Time{}},
		},
	}

	results := searchMemories(idx, "golang testing")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("expected mem1, got %s", results[0].MemoryID)
	}
	if results[0].Score == 0 {
		t.Errorf("expected non-zero score from word overlap fallback, got %v", results[0].Score)
	}
}

func TestSearchMemoriesTypeBoost(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 1, "mem2": 1},
		},
		DocFreq: map[string]int{
			"golang": 2,
		},
		DocCount: 2,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Name: "Project Memory", Description: "", Content: "golang", Type: "project", Created: time.Time{}},
			"mem2": {ID: "mem2", Name: "User Memory", Description: "", Content: "golang", Type: "user", Created: time.Time{}},
		},
	}

	results := searchMemories(idx, "golang")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("expected project memory mem1 to rank first, got %s", results[0].MemoryID)
	}
	if results[0].Score <= results[1].Score {
		t.Errorf("expected project memory score %v to be higher than user memory score %v", results[0].Score, results[1].Score)
	}
}

func TestSearchMemoriesRecencyBoost(t *testing.T) {
	now := time.Now()
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"mem1": 1, "mem2": 1},
		},
		DocFreq: map[string]int{
			"golang": 2,
		},
		DocCount: 2,
		Memories: map[string]Memory{
			"mem1": {ID: "mem1", Name: "Recent Memory", Description: "", Content: "golang", Type: "user", Created: now},
			"mem2": {ID: "mem2", Name: "Older Memory", Description: "", Content: "golang", Type: "user", Created: now.Add(-48 * time.Hour)},
		},
	}

	results := searchMemories(idx, "golang")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].MemoryID != "mem1" {
		t.Errorf("expected recent memory mem1 to rank first, got %s", results[0].MemoryID)
	}
	if results[0].Score <= results[1].Score {
		t.Errorf("expected recent memory score %v to be higher than older memory score %v", results[0].Score, results[1].Score)
	}
}

func TestBuildSearchIndexComputesDocFreqAcrossMemories(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang learning
type: user
created: 2026-07-11T12:00:00Z
---
Learn golang
`)

	writeMemoryFile(t, dir, "memory_2.md", `---
name: Memory Two
description: golang testing
type: project
created: 2026-07-11T13:00:00Z
---
Write tests in golang
`)

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 2 {
		t.Errorf("DocCount = %d, want 2", index.DocCount)
	}

	if index.DocFreq["golang"] != 2 {
		t.Errorf("DocFreq[golang] = %d, want 2", index.DocFreq["golang"])
	}

	if index.TermFreq["golang"]["memory_1"] != 2 {
		t.Errorf("TermFreq[golang][memory_1] = %d, want 2", index.TermFreq["golang"]["memory_1"])
	}
	if index.TermFreq["golang"]["memory_2"] != 2 {
		t.Errorf("TermFreq[golang][memory_2] = %d, want 2", index.TermFreq["golang"]["memory_2"])
	}
}

func TestExtractKeywordsTrimsPunctuationAndStops(t *testing.T) {
	got := extractKeywords("The a an of at Golang, testing! Rust.")
	want := []string{"golang", "testing", "rust"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractKeywords = %v, want %v", got, want)
	}
}

func TestExtractKeywordsIgnoresShortWords(t *testing.T) {
	got := extractKeywords("Go is a great language")
	want := []string{"great", "language"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractKeywords = %v, want %v", got, want)
	}
}

func TestCalculateTFIDFZeroDocCount(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{"golang": {"memory_1": 3}},
		DocFreq:  map[string]int{"golang": 1},
		DocCount: 0,
	}
	if got := calculateTFIDF(idx, "golang", "memory_1"); got != 0 {
		t.Errorf("calculateTFIDF = %v, want 0", got)
	}
}

func TestCalculateTFIDFScoring(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{"golang": {"memory_1": 3}},
		DocFreq:  map[string]int{"golang": 1},
		DocCount: 2,
	}
	got := calculateTFIDF(idx, "golang", "memory_1")
	want := 3 * math.Log(3.0 / 2.0)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("calculateTFIDF = %v, want %v", got, want)
	}
}

func TestCalculateTFIDFUnknownTerm(t *testing.T) {
	idx := &SearchIndex{
		TermFreq: map[string]map[string]int{},
		DocFreq:  map[string]int{},
		DocCount: 2,
	}
	if got := calculateTFIDF(idx, "missing", "memory_1"); got != 0 {
		t.Errorf("calculateTFIDF = %v, want 0", got)
	}
}

func TestExtractKeywordsEmptyString(t *testing.T) {
	got := extractKeywords("")
	if len(got) != 0 {
		t.Errorf("expected empty keywords, got %v", got)
	}
}

func TestSaveSearchIndexMarshalError(t *testing.T) {
	dir := t.TempDir()

	index := &SearchIndex{
		TermFreq: map[string]map[string]int{
			"golang": {"memory_1": 1},
		},
		DocFreq: map[string]int{
			"golang": 1,
		},
		DocCount: 1,
		Memories: map[string]Memory{
			"memory_1": {
				ID:          "memory_1",
				Name:        "Out of Range Memory",
				Description: "test",
				Type:        "user",
				Created:     time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
				Content:     "golang",
			},
		},
	}

	err := saveSearchIndex(dir, index)
	if err == nil {
		t.Fatal("expected saveSearchIndex to return an error")
	}
	if !strings.Contains(err.Error(), "MarshalJSON") {
		t.Errorf("expected JSON marshal error, got %v", err)
	}
}

func TestLoadSearchIndexFallsBackOnInvalidDateCache(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "search_index.json"), []byte(`{
  "TermFreq": {},
  "DocFreq": {},
  "DocCount": 1,
  "Memories": {
    "memory_1": {
      "id": "memory_1",
      "name": "Cached Memory",
      "description": "cached",
      "type": "user",
      "created": "not-a-date",
      "content": "cached content"
    }
  }
}`), 0644); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	index, err := loadSearchIndex(dir)
	if err != nil {
		t.Fatalf("loadSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	if index.DocCount != 1 {
		t.Errorf("DocCount = %d, want 1", index.DocCount)
	}
	if _, ok := index.Memories["memory_1"]; !ok {
		t.Errorf("expected memory_1 to be indexed from rebuild, got %v", index.Memories)
	}
	if index.Memories["memory_1"].Name != "Memory One" {
		t.Errorf("Name = %q, want %q", index.Memories["memory_1"].Name, "Memory One")
	}
}

func TestSearchMemoriesStopWordsOnly(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
Write tests in golang
`)

	writeMemoryFile(t, dir, "memory_2.md", `---
name: Memory Two
description: rust project
type: user
created: 2026-07-11T10:00:00Z
---
rust programming
`)

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	results := searchMemories(index, "the a an of at")
	if len(results) != 0 {
		t.Errorf("expected 0 results for stop-word-only query, got %d", len(results))
	}
}

func TestSearchMemoriesExactPhraseCaseInsensitive(t *testing.T) {
	dir := t.TempDir()

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: 2026-07-11T12:00:00Z
---
golang testing project
`)

	writeMemoryFile(t, dir, "memory_2.md", `---
name: Memory Two
description: rust project
type: user
created: 2026-07-11T12:00:00Z
---
GOLANG TESTING rust
`)

	index, err := buildSearchIndex(dir)
	if err != nil {
		t.Fatalf("buildSearchIndex failed: %v", err)
	}
	if index == nil {
		t.Fatal("expected index, got nil")
	}

	results := searchMemories(index, "GOLANG TESTING")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].MemoryID != "memory_1" {
		t.Errorf("expected project memory memory_1 to rank first, got %s", results[0].MemoryID)
	}
	if results[0].Score <= results[1].Score {
		t.Errorf("expected memory_1 score %v to be higher than memory_2 score %v", results[0].Score, results[1].Score)
	}
}
