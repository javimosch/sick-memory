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

func TestHandleDeleteJSON(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.WriteFile(filepath.Join(dir, "memory_123.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	os.Args = []string{"cmd", "delete", "123"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleDelete(cfg)
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
	if data["status"] != "deleted" {
		t.Errorf("status = %v, want %q", data["status"], "deleted")
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_123.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}

func TestHandleDeleteTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.WriteFile(filepath.Join(dir, "memory_123.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	os.Args = []string{"cmd", "delete", "123"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleDelete(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Memory 123 deleted successfully") {
		t.Errorf("expected deleted message, got %q", got)
	}
}

func TestHandleDeleteMissingArgument(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		os.Args = []string{"cmd", "delete"}
		handleDelete(&Config{MemoryDir: ""})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleDeleteMissingArgument", "-test.v")
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

func TestHandleDeleteMemoryNotFound(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		os.Args = []string{"cmd", "delete", "999"}
		handleDelete(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleDeleteMemoryNotFound", "-test.v")
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

func TestHandleDeleteReadDirError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		os.Args = []string{"cmd", "delete", "123"}
		handleDelete(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleDeleteReadDirError", "-test.v")
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

func TestHandleDeleteRemoveError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		memoryDir := filepath.Join(dir, "memory_123")
		if err := os.MkdirAll(memoryDir, 0755); err != nil {
			t.Fatalf("failed to create memory directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(memoryDir, "keep.txt"), []byte("keep"), 0644); err != nil {
			t.Fatalf("failed to create file inside memory directory: %v", err)
		}
		os.Args = []string{"cmd", "delete", "123"}
		handleDelete(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleDeleteRemoveError", "-test.v")
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

func TestMainDeleteMissingArgument(t *testing.T) {
	if os.Getenv("MAIN_DELETE_MISSING") == "1" {
		os.Args = []string{"sick-memory", "delete"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainDeleteMissingArgument", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_DELETE_MISSING=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 80 {
		t.Fatalf("expected exit code 80, got %v", err)
	}

	if !strings.Contains(string(out), "Memory ID required for delete") {
		t.Errorf("expected missing argument message, got:\n%s", out)
	}
}

func TestMainDelete(t *testing.T) {
	if os.Getenv("MAIN_DELETE") == "1" {
		os.Args = []string{"sick-memory", "delete", "123"}
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

	writeMemoryFile(t, memDir, "memory_123.md", "content")

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainDelete$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_DELETE=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainDelete subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memory 123 deleted successfully") {
		t.Errorf("expected deleted message, got:\n%s", got)
	}

	if _, err := os.Stat(filepath.Join(memDir, "memory_123.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}

func TestMainDeleteJSON(t *testing.T) {
	if os.Getenv("MAIN_DELETE_JSON") == "1" {
		os.Args = []string{"sick-memory", "delete", "123", "--json"}
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

	writeMemoryFile(t, memDir, "memory_123.md", "content")

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainDeleteJSON$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_DELETE_JSON=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainDeleteJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"id": "123"`) {
		t.Errorf("expected id in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"status": "deleted"`) {
		t.Errorf("expected status deleted in output, got:\n%s", got)
	}

	if _, err := os.Stat(filepath.Join(memDir, "memory_123.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}
