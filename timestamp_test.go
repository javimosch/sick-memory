package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreatedTimestampFormat(t *testing.T) {
	now := time.Now()
	created := now.Format(time.RFC3339)

	parsed, err := time.Parse(time.RFC3339, created)
	if err != nil {
		t.Fatalf("Failed to parse RFC3339 timestamp: %v", err)
	}
	if parsed.Unix() != now.Unix() {
		t.Errorf("Timestamp mismatch: got %d, want %d", parsed.Unix(), now.Unix())
	}
}

func TestParseMemoryCreatedRFC3339(t *testing.T) {
	content := `---
name: Test Memory
description: test description
type: user
created: 2026-06-17T04:30:00Z
---

Test content
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_memory.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	memory := parseMemory(content, filepath.Base(filePath))

	expected := time.Date(2026, 6, 17, 4, 30, 0, 0, time.UTC)
	if !memory.Created.Equal(expected) {
		t.Errorf("Created timestamp mismatch: got %v, want %v", memory.Created, expected)
	}
}

func TestHandleRememberWritesRFC3339(t *testing.T) {
	// Save and override global state
	origArgs := os.Args
	origNoInteractive := noInteractive
	defer func() {
		os.Args = origArgs
		noInteractive = origNoInteractive
	}()

	noInteractive = true

	// Set up a temp memory directory
	tmpDir := t.TempDir()

	// Simulate: sick-memory remember "hello world" --memory-dir <tmpDir>
	os.Args = []string{
		"sick-memory",
		"remember",
		"hello world",
		"--memory-dir",
		tmpDir,
	}

	cfg := &Config{
		MemoryDir: tmpDir,
		GlobalConfig: GlobalConfig{
			AutoIndex: false,
		},
	}

	handleRemember(cfg)

	// Verify a memory file was created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read memory dir: %v", err)
	}

	var found bool
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(tmpDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", entry.Name(), err)
		}

		content := string(data)
		if !strings.Contains(content, "created:") {
			continue
		}

		// Parse the memory to verify the timestamp
		memory := parseMemory(content, entry.Name())
		if memory.Created.IsZero() {
			t.Errorf("Memory %s has zero created timestamp", entry.Name())
			continue
		}

		// Verify it's valid RFC3339 by round-tripping
		rfc3339Str := memory.Created.Format(time.RFC3339)
		parsed, err := time.Parse(time.RFC3339, rfc3339Str)
		if err != nil {
			t.Errorf("Created timestamp %v does not round-trip through RFC3339: %v", memory.Created, err)
			continue
		}
		if !parsed.Equal(memory.Created) {
			t.Errorf("Created timestamp %v does not match parsed %v", memory.Created, parsed)
		}

		found = true
	}

	if !found {
		t.Fatal("No memory file with created timestamp was generated")
	}
}

func TestMemoryIDUsesUnixNano(t *testing.T) {
	// Verify that two rapid handleRemember calls produce different filenames
	origArgs := os.Args
	origNoInteractive := noInteractive
	defer func() {
		os.Args = origArgs
		noInteractive = origNoInteractive
	}()

	noInteractive = true

	files := make(map[string]bool)

	for i := 0; i < 5; i++ {
		tmpDir := t.TempDir()
		os.Args = []string{
			"sick-memory",
			"remember",
			"content",
			"--memory-dir",
			tmpDir,
		}
		cfg := &Config{
			MemoryDir: tmpDir,
			GlobalConfig: GlobalConfig{
				AutoIndex: false,
			},
		}
		handleRemember(cfg)

		entries, _ := os.ReadDir(tmpDir)
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".md" && e.Name() != "MEMORY.md" {
				if files[e.Name()] {
					t.Errorf("Duplicate filename generated: %s", e.Name())
				}
				files[e.Name()] = true
			}
		}
	}

	if len(files) < 5 {
		t.Errorf("Expected 5 unique memory files, got %d", len(files))
	}
}
