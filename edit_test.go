package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleEditJSONTruncatesLongDescription(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	writeMemoryFile(t, dir, "memory_123.md", `---
name: Memory 123
description: original content
type: user
created: 2026-07-11T12:00:00Z
---

Original content
`)

	os.Args = []string{"cmd", "edit", "123", "This is a very long updated description that should be truncated"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleEdit(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["id"] != "123" {
		t.Errorf("id = %v, want %q", data["id"], "123")
	}
	if data["status"] != "updated" {
		t.Errorf("status = %v, want %q", data["status"], "updated")
	}

	want := "This is a very long updated description that shoul..."
	if data["description"] != want {
		t.Errorf("description = %v, want %q", data["description"], want)
	}

	updated, err := os.ReadFile(filepath.Join(dir, "memory_123.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "description: "+want) {
		t.Errorf("expected truncated description in frontmatter, got %q", string(updated))
	}
}

func TestHandleEditJSON(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	writeMemoryFile(t, dir, "memory_123.md", `---
name: Memory 123
description: original content
type: user
created: 2026-07-11T12:00:00Z
---

Original content
`)

	os.Args = []string{"cmd", "edit", "123", "Updated content for golang tests"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleEdit(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["id"] != "123" {
		t.Errorf("id = %v, want %q", data["id"], "123")
	}
	if data["status"] != "updated" {
		t.Errorf("status = %v, want %q", data["status"], "updated")
	}

	updated, err := os.ReadFile(filepath.Join(dir, "memory_123.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "Updated content for golang tests") {
		t.Errorf("expected updated content in file, got %q", string(updated))
	}
	if !strings.Contains(string(updated), "description: Updated content for golang tests") {
		t.Errorf("expected updated description in frontmatter, got %q", string(updated))
	}
}

func TestHandleEditTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	writeMemoryFile(t, dir, "memory_456.md", `---
name: Memory 456
description: original
type: project
created: 2026-07-11T12:00:00Z
---

Original content
`)

	os.Args = []string{"cmd", "edit", "456", "new content"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleEdit(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Memory 456 updated successfully") {
		t.Errorf("expected updated message, got %q", got)
	}
}

func TestHandleEditJSONTruncatesAtNewline(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	writeMemoryFile(t, dir, "memory_123.md", `---
name: Memory 123
description: original content
type: user
created: 2026-07-11T12:00:00Z
---

Original content
`)

	os.Args = []string{"cmd", "edit", "123", "First line\nsecond line"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleEdit(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}

	want := "First line"
	if data["description"] != want {
		t.Errorf("description = %v, want %q", data["description"], want)
	}

	updated, err := os.ReadFile(filepath.Join(dir, "memory_123.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "description: "+want) {
		t.Errorf("expected truncated description in frontmatter, got %q", string(updated))
	}
	if !strings.Contains(string(updated), "First line\nsecond line") {
		t.Errorf("expected full content after frontmatter, got %q", string(updated))
	}
}

func TestHandleEditMissingArgument(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		os.Args = []string{"cmd", "edit", "123"}
		handleEdit(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleEditMissingArgument", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 80 {
		t.Errorf("expected exit code 80, got %d", exitErr.ExitCode())
	}
}
