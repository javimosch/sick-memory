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

func TestEditMainTextOutput(t *testing.T) {
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
	os.Args = []string{"sick-memory", "edit", "1", "new main text"}
	jsonOutput = false
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

	got := string(out)
	if !strings.Contains(got, "Memory 1 updated successfully") {
		t.Errorf("expected updated message, got %q", got)
	}

	updated, err := os.ReadFile(filepath.Join(dir, "memory_1.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "new main text") {
		t.Errorf("expected updated content in file, got:\n%s", string(updated))
	}
}

func TestEditMainMemoryNotFound(t *testing.T) {
	if os.Getenv("EDIT_MAIN_MEMORY_NOT_FOUND") == "1" {
		home := os.Getenv("HOME")
		repo := os.Getenv("REPO_ROOT")
		memDir := filepath.Join(home, ".sick-memory", "projects", sanitizePath(repo), "memory")
		if err := os.MkdirAll(memDir, 0755); err != nil {
			t.Fatalf("failed to create memory dir: %v", err)
		}
		os.Args = []string{"sick-memory", "edit", "999", "new content"}
		main()
		return
	}

	home := t.TempDir()
	repo := t.TempDir()
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestEditMainMemoryNotFound$", "-test.v")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "EDIT_MAIN_MEMORY_NOT_FOUND=1", "HOME="+home, "REPO_ROOT="+repo)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 92 {
		t.Fatalf("expected exit code 92, got %v", err)
	}

	if !strings.Contains(string(out), "Memory with ID 999 not found") {
		t.Errorf("expected memory not found message, got:\n%s", out)
	}
}

func TestEditMainMissingArgument(t *testing.T) {
	if os.Getenv("EDIT_MAIN_MISSING") == "1" {
		dir := t.TempDir()
		t.Setenv("HOME", dir)
		os.Args = []string{"sick-memory", "edit"}
		memoryDir = ""
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestEditMainMissingArgument$", "-test.v")
	cmd.Env = append(os.Environ(), "EDIT_MAIN_MISSING=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 80 {
		t.Fatalf("expected exit code 80, got %v", err)
	}

	if !strings.Contains(string(out), "Memory ID and new content required") {
		t.Errorf("expected missing argument message, got:\n%s", out)
	}
}

func TestEditMainReadDirError(t *testing.T) {
	if os.Getenv("EDIT_MAIN_READ_DIR") == "1" {
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		memoryDir = dir
		jsonOutput = true
		os.Args = []string{"sick-memory", "edit", "1", "new content"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestEditMainReadDirError$", "-test.v")
	cmd.Env = append(os.Environ(), "EDIT_MAIN_READ_DIR=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 92 {
		t.Fatalf("expected exit code 92, got %v", err)
	}

	if !strings.Contains(string(out), "Cannot read memory directory") {
		t.Errorf("expected read directory error message, got:\n%s", out)
	}
}

func TestEditMainReadFileError(t *testing.T) {
	if os.Getenv("EDIT_MAIN_READ_FILE") == "1" {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "memory_1.md"), 0755); err != nil {
			t.Fatalf("failed to create blocking memory directory: %v", err)
		}
		memoryDir = dir
		jsonOutput = true
		os.Args = []string{"sick-memory", "edit", "1", "new content"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestEditMainReadFileError$", "-test.v")
	cmd.Env = append(os.Environ(), "EDIT_MAIN_READ_FILE=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 110 {
		t.Fatalf("expected exit code 110, got %v", err)
	}

	if !strings.Contains(string(out), "Cannot read memory file") {
		t.Errorf("expected read file error message, got:\n%s", out)
	}
}

func TestEditMainTruncatesLongDescription(t *testing.T) {
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
	os.Args = []string{"sick-memory", "edit", "1", "This is a very long updated description that should be truncated"}
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

	want := "This is a very long updated description that shoul..."
	if data["description"] != want {
		t.Errorf("description = %v, want %q", data["description"], want)
	}

	updated, err := os.ReadFile(filepath.Join(dir, "memory_1.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "description: "+want) {
		t.Errorf("expected truncated description in frontmatter, got:\n%s", string(updated))
	}
	if !strings.Contains(string(updated), "This is a very long updated description that should be truncated") {
		t.Errorf("expected full content after frontmatter, got:\n%s", string(updated))
	}
}
