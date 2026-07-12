package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditMain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: original
type: user
created: 2026-07-11T12:00:00Z
---

Original content
`)

	oldArgs := os.Args
	oldJSON := jsonOutput
	oldMemoryDir := memoryDir
	oldNoInteractive := noInteractive
	defer func() {
		os.Args = oldArgs
		jsonOutput = oldJSON
		memoryDir = oldMemoryDir
		noInteractive = oldNoInteractive
	}()
	os.Args = []string{"sick-memory", "edit", "1", "new main content"}
	jsonOutput = true
	memoryDir = dir
	noInteractive = false

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	main()
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "1" {
		t.Errorf("id = %v, want %q", data["id"], "1")
	}
	if data["status"] != "updated" {
		t.Errorf("status = %v, want %q", data["status"], "updated")
	}
	if data["description"] != "new main content" {
		t.Errorf("description = %v, want %q", data["description"], "new main content")
	}

	updated, err := os.ReadFile(filepath.Join(dir, "memory_1.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "new main content") {
		t.Errorf("expected updated content in file, got:\n%s", string(updated))
	}
}
