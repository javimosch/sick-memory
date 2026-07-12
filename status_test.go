package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleStatusUninitialized(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: filepath.Join(dir, "does-not-exist")}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
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
	if data["status"] != "uninitialized" {
		t.Errorf("expected status 'uninitialized', got %v", data["status"])
	}
}

func TestHandleStatusActive(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
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
	if data["status"] != "active" {
		t.Errorf("expected status 'active', got %v", data["status"])
	}
	if data["path"] != dir {
		t.Errorf("expected path %q, got %v", dir, data["path"])
	}
	if count, ok := data["count"].(float64); !ok || count != 0 {
		t.Errorf("expected count 0, got %v", data["count"])
	}
}

func TestHandleStatusUninitializedTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: filepath.Join(dir, "does-not-exist")}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system status: uninitialized") {
		t.Errorf("expected uninitialized status, got %q", got)
	}
	if !strings.Contains(got, "Run 'sick-memory init' to initialize.") {
		t.Errorf("expected init hint, got %q", got)
	}
}

func TestHandleStatusActiveTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system status: active") {
		t.Errorf("expected active status, got %q", got)
	}
	if !strings.Contains(got, "Memory directory:") {
		t.Errorf("expected memory directory, got %q", got)
	}
	if !strings.Contains(got, "Total memories: 0") {
		t.Errorf("expected total memories, got %q", got)
	}
}

func TestHandleStatusActiveWithMemories(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("not a memory"), 0644); err != nil {
		t.Fatalf("failed to write README file: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
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
	if data["status"] != "active" {
		t.Errorf("expected status 'active', got %v", data["status"])
	}
	if count, ok := data["count"].(float64); !ok || count != 1 {
		t.Errorf("expected count 1, got %v", data["count"])
	}
	if data["path"] != dir {
		t.Errorf("expected path %q, got %v", dir, data["path"])
	}
}
