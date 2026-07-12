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
