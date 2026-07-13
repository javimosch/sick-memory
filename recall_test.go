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

func TestHandleRecallJSONNoResults(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	results, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
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

func TestHandleRecallJSONMultipleResults(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}}

	recent := time.Now().UTC().Format(time.RFC3339)
	older := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: user
created: `+recent+`
---

Write tests in golang
`)

	writeMemoryFile(t, dir, "memory_2.md", `---
name: Memory Two
description: golang code
type: user
created: `+older+`
---

Build a golang project
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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	results, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	first, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T", results[0])
	}
	if first["memory_id"] != "memory_1" {
		t.Errorf("first result memory_id = %v, want %q", first["memory_id"], "memory_1")
	}

	second, ok := results[1].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T", results[1])
	}
	if second["memory_id"] != "memory_2" {
		t.Errorf("second result memory_id = %v, want %q", second["memory_id"], "memory_2")
	}
}

func TestHandleRecallAllMemoriesSortedByRecency(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}}

	recent := time.Now().UTC().Format(time.RFC3339)
	older := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)

	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: recent memory
type: user
created: `+recent+`
---

Recent content
`)

	writeMemoryFile(t, dir, "memory_2.md", `---
name: Memory Two
description: older memory
type: user
created: `+older+`
---

Older content
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
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	first, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T", results[0])
	}
	if first["memory_id"] != "memory_1" {
		t.Errorf("first result memory_id = %v, want %q", first["memory_id"], "memory_1")
	}

	second, ok := results[1].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T", results[1])
	}
	if second["memory_id"] != "memory_2" {
		t.Errorf("second result memory_id = %v, want %q", second["memory_id"], "memory_2")
	}
}

func TestHandleRecallMissingDirectory(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		handleRecall(&Config{MemoryDir: dir, GlobalConfig: GlobalConfig{AutoIndex: false}})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleRecallMissingDirectory$", "-test.v")
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

func TestHandleRecallSkipsGlobalFlags(t *testing.T) {
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

	os.Args = []string{"cmd", "recall", "--json", "--memory-dir=/tmp", "golang"}

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
	if result["memory_type"] != "user" {
		t.Errorf("memory_type = %v, want %q", result["memory_type"], "user")
	}
}

func TestMainRecallJSON(t *testing.T) {
	if os.Getenv("MAIN_RECALL_JSON") == "1" {
		os.Args = []string{"sick-memory", "recall", "--json", "golang"}
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

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, memDir, "memory_1.md", `---
name: Memory One
description: golang search
type: user
created: `+created+`
---

main recall test
`)

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainRecallJSON$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_RECALL_JSON=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainRecallJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"memory_id":"memory_1"`) {
		t.Errorf("expected memory_id in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}
}

func TestMainSearchAlias(t *testing.T) {
	if os.Getenv("MAIN_SEARCH_ALIAS") == "1" {
		os.Args = []string{"sick-memory", "search", "--json", "golang"}
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

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, memDir, "memory_1.md", `---
name: Memory One
description: golang search
type: user
created: `+created+`
---

main search alias test
`)

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainSearchAlias$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_SEARCH_ALIAS=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainSearchAlias subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"memory_id":"memory_1"`) {
		t.Errorf("expected memory_id in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}
}

func TestMainRecallText(t *testing.T) {
	if os.Getenv("MAIN_RECALL_TEXT") == "1" {
		os.Args = []string{"sick-memory", "recall", "golang"}
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

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, memDir, "memory_1.md", `---
name: Memory One
description: golang search
type: user
created: `+created+`
---

main recall text test
`)

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainRecallText$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_RECALL_TEXT=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainRecallText subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Found 1 memories matching: golang") {
		t.Errorf("expected match summary, got:\n%s", got)
	}
	if !strings.Contains(got, "ID: memory_1") {
		t.Errorf("expected memory ID in output, got:\n%s", got)
	}
}

func TestMainRecallNoQueryText(t *testing.T) {
	if os.Getenv("MAIN_RECALL_NO_QUERY_TEXT") == "1" {
		os.Args = []string{"sick-memory", "recall"}
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

	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, memDir, "memory_1.md", `---
name: Memory One
description: all memories
type: user
created: `+created+`
---

main recall no query text test
`)

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainRecallNoQueryText$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_RECALL_NO_QUERY_TEXT=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainRecallNoQueryText subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "All memories in") {
		t.Errorf("expected all memories header, got:\n%s", got)
	}
	if !strings.Contains(got, "ID: memory_1") {
		t.Errorf("expected memory ID in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Total memories: 1") {
		t.Errorf("expected total memories count, got:\n%s", got)
	}
}

func TestRecallMain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang testing
type: project
created: `+created+`
---

Write tests in golang
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
	os.Args = []string{"sick-memory", "recall", "--json", "golang"}
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
	if !ok || len(results) == 0 {
		t.Fatalf("expected non-empty results, got %T", resp.Data)
	}
	first, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", results[0])
	}
	if first["memory_id"] != "memory_1" {
		t.Errorf("memory_id = %v, want %q", first["memory_id"], "memory_1")
	}
}

func TestSearchMain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	created := time.Now().UTC().Format(time.RFC3339)
	writeMemoryFile(t, dir, "memory_1.md", `---
name: Memory One
description: golang search
type: user
created: `+created+`
---

main search alias test
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
	os.Args = []string{"sick-memory", "search", "--json", "golang"}
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
	if !ok || len(results) == 0 {
		t.Fatalf("expected non-empty results, got %T", resp.Data)
	}
	first, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", results[0])
	}
	if first["memory_id"] != "memory_1" {
		t.Errorf("memory_id = %v, want %q", first["memory_id"], "memory_1")
	}
}
