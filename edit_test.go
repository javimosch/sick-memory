package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleEditMissingArgument$", "-test.v")
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

func TestHandleEditReadFileError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "memory_123.md"), 0755); err != nil {
			t.Fatalf("failed to create memory directory: %v", err)
		}
		os.Args = []string{"cmd", "edit", "123", "new"}
		handleEdit(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleEditReadFileError$", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 110 {
		t.Errorf("expected exit code 110, got %d", exitErr.ExitCode())
	}
}

func TestHandleEditMemoryNotFound(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		os.Args = []string{"cmd", "edit", "999", "new content"}
		handleEdit(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleEditMemoryNotFound$", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}

func TestHandleEditReadDirError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		os.Args = []string{"cmd", "edit", "123", "new content"}
		handleEdit(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleEditReadDirError$", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}

func TestHandleEditWriteError(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires /proc/self/environ")
	}

	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		if err := os.Symlink("/proc/self/environ", filepath.Join(dir, "memory_123.md")); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}
		os.Args = []string{"cmd", "edit", "123", "new content"}
		handleEdit(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleEditWriteError$", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 110 {
		t.Errorf("expected exit code 110, got %d", exitErr.ExitCode())
	}
}

func TestMainEditText(t *testing.T) {
	if os.Getenv("MAIN_EDIT_TEXT") == "1" {
		os.Args = []string{"sick-memory", "edit", "123", "main updated content"}
		main()
		return
	}

	home := t.TempDir()
	repo := t.TempDir()
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	memDir := filepath.Join(home, ".sick-memory", "projects", sanitizePath(repo), "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	writeMemoryFile(t, memDir, "memory_123.md", `---
name: Memory 123
description: original
type: user
created: 2026-07-11T12:00:00Z
---

Original content
`)

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainEditText$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_EDIT_TEXT=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainEditText subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memory 123 updated successfully") {
		t.Errorf("expected updated message, got:\n%s", got)
	}

	updated, err := os.ReadFile(filepath.Join(memDir, "memory_123.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "main updated content") {
		t.Errorf("expected updated content in file, got:\n%s", string(updated))
	}
}

func TestMainEditJSON(t *testing.T) {
	if os.Getenv("MAIN_EDIT_JSON") == "1" {
		os.Args = []string{"sick-memory", "edit", "123", "main updated content", "--json"}
		main()
		return
	}

	home := t.TempDir()
	repo := t.TempDir()
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	memDir := filepath.Join(home, ".sick-memory", "projects", sanitizePath(repo), "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	writeMemoryFile(t, memDir, "memory_123.md", `---
name: Memory 123
description: original
type: user
created: 2026-07-11T12:00:00Z
---

Original content
`)

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainEditJSON$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_EDIT_JSON=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainEditJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"status": "updated"`) {
		t.Errorf("expected updated status in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"description": "main updated content`) {
		t.Errorf("expected updated description in output, got:\n%s", got)
	}

	updated, err := os.ReadFile(filepath.Join(memDir, "memory_123.md"))
	if err != nil {
		t.Fatalf("failed to read updated memory file: %v", err)
	}
	if !strings.Contains(string(updated), "main updated content --json") {
		t.Errorf("expected updated content in file, got:\n%s", string(updated))
	}
}
