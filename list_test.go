package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleListEmptyDirectory(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleList(cfg)
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

	ids, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 memory IDs, got %d", len(ids))
	}
}

func TestHandleListWithMemories(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "memory_2.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleList(cfg)
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

	ids, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 memory IDs, got %d", len(ids))
	}

	if ids[0] != "memory_1.md" && ids[0] != "memory_2.md" {
		t.Errorf("unexpected memory ID %q", ids[0])
	}
}
