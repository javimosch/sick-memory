package main

import (
	"os"
	"path/filepath"
	"reflect"
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

func writeMemoryFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}
