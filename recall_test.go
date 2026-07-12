package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHandleRecallJSONWithResults(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}}

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: `+created+`
---

Write tests in golang
`)

	os.Args = []string{"cmd", "recall", "golang", "testing"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRecall(cfg)
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

	results, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T", results[0])
	}
	if result["memory_id"] != "memory_1" {
		t.Errorf("memory_id = %v, want %q", result["memory_id"], "memory_1")
	}
	if result["memory_type"] != "project" {
		t.Errorf("memory_type = %v, want %q", result["memory_type"], "project")
	}
	if !strings.Contains(result["content"].(string), "golang") {
		t.Errorf("expected content to contain 'golang', got %q", result["content"])
	}
}

func TestHandleRecallTextNoResults(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}}

	os.Args = []string{"cmd", "recall", "rust"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRecall(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "No memories found matching: rust") {
		t.Errorf("expected no results message, got %q", got)
	}
}

func TestHandleRecallTextWithResults(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}}

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: `+created+`
---

Write tests in golang
`)

	os.Args = []string{"cmd", "recall", "golang"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRecall(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Found 1 memories matching: golang") {
		t.Errorf("expected match summary, got %q", got)
	}
	if !strings.Contains(got, "ID: memory_1") {
		t.Errorf("expected memory ID in output, got %q", got)
	}
	if !strings.Contains(got, "golang") {
		t.Errorf("expected content to contain 'golang', got %q", got)
	}
}

func TestHandleRecallJSONAllMemories(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}}

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: user
created: `+created+`
---

Write tests in golang
`)

	os.Args = []string{"cmd", "recall"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRecall(cfg)
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

	results, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T", results[0])
	}
	if result["memory_id"] != "memory_1" {
		t.Errorf("memory_id = %v, want %q", result["memory_id"], "memory_1")
	}
}

func TestHandleRecallTextAllMemories(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}}

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: user
created: `+created+`
---

Write tests in golang
`)

	os.Args = []string{"cmd", "recall"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRecall(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "All memories in") {
		t.Errorf("expected all memories header, got %q", got)
	}
	if !strings.Contains(got, "ID: memory_1") {
		t.Errorf("expected memory ID in output, got %q", got)
	}
	if !strings.Contains(got, "Total memories: 1") {
		t.Errorf("expected total memories count, got %q", got)
	}
}

