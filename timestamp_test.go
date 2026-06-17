package main

import (
	"os"
	"path/filepath"
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
	now := time.Now()

	generated := `---
name: Memory test_memory
description: test content
type: user
created: ` + now.Format(time.RFC3339) + `
---

test content
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_generated.md")
	if err := os.WriteFile(filePath, []byte(generated), 0644); err != nil {
		t.Fatal(err)
	}

	memory := parseMemory(generated, filepath.Base(filePath))

	if memory.Created.IsZero() {
		t.Error("Created timestamp should not be zero after parsing RFC3339")
	}
	if memory.Created.Unix() != now.Unix() {
		t.Errorf("Created timestamp mismatch: got %v, want %v", memory.Created.Unix(), now.Unix())
	}
}
