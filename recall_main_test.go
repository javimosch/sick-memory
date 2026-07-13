package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRecallMainMemoryDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang recall
type: project
created: `+created+`
---
main recall --memory-dir test
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
	os.Args = []string{"sick-memory", "recall", "--json", "golang", "--memory-dir", dir}
	jsonOutput = false
	memoryDir = ""
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

	data, ok := resp["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Fatalf("expected non-empty data, got %T", resp["data"])
	}
	first, ok := data[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", data[0])
	}
	if first["memory_id"] != "memory_1" {
		t.Errorf("memory_id = %v, want %q", first["memory_id"], "memory_1")
	}
}

func TestRecallMainMemoryDirTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang recall
type: project
created: `+created+`
---
main recall --memory-dir text
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
	os.Args = []string{"sick-memory", "recall", "golang", "--memory-dir", dir}
	jsonOutput = false
	memoryDir = ""
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
	if !strings.Contains(got, "Found 1 memories matching: golang") {
		t.Errorf("expected match summary, got %q", got)
	}
	if !strings.Contains(got, "ID: memory_1") {
		t.Errorf("expected memory ID in output, got %q", got)
	}
}

func TestRecallMainMissingDirectory(t *testing.T) {
	if os.Getenv("RECALL_MAIN_MISSING_DIR") == "1" {
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		jsonOutput = true
		memoryDir = dir
		os.Args = []string{"sick-memory", "recall", "--json", "golang"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestRecallMainMissingDirectory$", "-test.v")
	cmd.Env = append(os.Environ(), "RECALL_MAIN_MISSING_DIR=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil\n%s", out)
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 92 {
		t.Fatalf("expected exit code 92, got %v", err)
	}

	if !strings.Contains(string(out), "Memory directory not found") {
		t.Errorf("expected memory directory not found message, got:\n%s", out)
	}
}

func TestRecallMainNoQuery(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: recall all
type: project
created: `+created+`
---
main recall all memories
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
	os.Args = []string{"sick-memory", "recall", "--json"}
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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	results, ok := resp.Data.([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("expected 1 result, got %T %v", resp.Data, resp.Data)
	}
}

func TestRecallMainNoQueryTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: recall all
type: project
created: `+created+`
---
main recall all memories
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
	os.Args = []string{"sick-memory", "recall"}
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

func TestSearchMainMemoryDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang search
type: project
created: `+created+`
---
main search memory dir test
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
	os.Args = []string{"sick-memory", "search", "--json", "golang", "--memory-dir", dir}
	jsonOutput = false
	memoryDir = ""
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

	data, ok := resp["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Fatalf("expected non-empty data, got %T", resp["data"])
	}
	first, ok := data[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", data[0])
	}
	if first["memory_id"] != "memory_1" {
		t.Errorf("memory_id = %v, want %q", first["memory_id"], "memory_1")
	}
}

func TestSearchMainMemoryDirTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang search
type: project
created: `+created+`
---
main search memory dir text
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
	os.Args = []string{"sick-memory", "search", "golang", "--memory-dir", dir}
	jsonOutput = false
	memoryDir = ""
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
	if !strings.Contains(got, "Found 1 memories matching: golang") {
		t.Errorf("expected match summary, got %q", got)
	}
	if !strings.Contains(got, "ID: memory_1") {
		t.Errorf("expected memory ID in output, got %q", got)
	}
}

func TestSearchMainNoQuery(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: search all
type: project
created: `+created+`
---
main search all memories
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
	os.Args = []string{"sick-memory", "search", "--json"}
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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	results, ok := resp.Data.([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("expected 1 result, got %T %v", resp.Data, resp.Data)
	}
}

func TestSearchMainNoQueryTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: search all text
type: project
created: `+created+`
---
main search all memories text
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
	os.Args = []string{"sick-memory", "search"}
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
