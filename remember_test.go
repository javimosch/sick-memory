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

func TestHandleRememberJSON(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: false},
	}

	os.Args = []string{"cmd", "remember", "test memory"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_"+id+".md")); err != nil {
		t.Errorf("expected memory file to exist: %v", err)
	}
}

func TestHandleRememberTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: false},
	}

	os.Args = []string{"cmd", "remember", "test memory"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.HasPrefix(got, "Memory saved with ID:") {
		t.Errorf("expected saved message, got %q", got)
	}
}

func TestHandleRememberAutoIndex(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: true},
	}

	os.Args = []string{"cmd", "remember", "auto index test"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}

	if _, err := os.Stat(filepath.Join(dir, "search_index.json")); err != nil {
		t.Errorf("expected search index to be written: %v", err)
	}
}

func TestHandleRememberMissingContentNoInteractive(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		noInteractive = true
		os.Args = []string{"cmd", "remember"}
		handleRemember(&Config{MemoryDir: t.TempDir()})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleRememberMissingContentNoInteractive", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 85 {
		t.Errorf("expected exit code 85, got %d", exitErr.ExitCode())
	}
}

func TestHandleRememberMissingContentInteractive(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		noInteractive = false
		os.Args = []string{"cmd", "remember"}
		handleRemember(&Config{MemoryDir: t.TempDir()})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleRememberMissingContentInteractive", "-test.v")
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
