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

func TestHandleListTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
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

	got := string(out)
	if !strings.Contains(got, "Memories in") {
		t.Errorf("expected output to contain 'Memories in', got %q", got)
	}
	if !strings.Contains(got, "memory_1.md") {
		t.Errorf("expected output to contain memory ID, got %q", got)
	}
	if !strings.Contains(got, "Total memories: 1") {
		t.Errorf("expected output to contain 'Total memories: 1', got %q", got)
	}
}

func TestHandleListTextOutputEmpty(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
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

	got := string(out)
	if !strings.Contains(got, "Memories in") {
		t.Errorf("expected output to contain 'Memories in', got %q", got)
	}
	if !strings.Contains(got, "Total memories: 0") {
		t.Errorf("expected output to contain 'Total memories: 0', got %q", got)
	}
	if strings.Contains(got, "memory_") {
		t.Errorf("expected no memory IDs in empty output, got %q", got)
	}
}

func TestHandleListMissingDirectoryJSON(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		handleList(&Config{MemoryDir: filepath.Join(t.TempDir(), "does-not-exist")})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleListMissingDirectoryJSON", "-test.v")
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

func TestHandleListMissingDirectoryText(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = false
		handleList(&Config{MemoryDir: filepath.Join(t.TempDir(), "does-not-exist")})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleListMissingDirectoryText", "-test.v")
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

func TestMainListMissingDirectory(t *testing.T) {
	if os.Getenv("MAIN_LIST_MISSING") == "1" {
		os.Args = []string{"sick-memory", "list", "--json", "--memory-dir", filepath.Join(t.TempDir(), "does-not-exist")}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainListMissingDirectory", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_LIST_MISSING=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 92 {
		t.Fatalf("expected exit code 92, got %v", err)
	}

	if !strings.Contains(string(out), "Memory directory not found") {
		t.Errorf("expected missing directory message, got:\n%s", out)
	}
}

func TestMainLsAlias(t *testing.T) {
	if os.Getenv("MAIN_LS_ALIAS") == "1" {
		os.Args = []string{"sick-memory", "ls", "--json"}
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

	writeMemoryFile(t, memDir, "memory_1.md", "content")

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainLsAlias$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_LS_ALIAS=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainLsAlias subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"memory_1.md"`) {
		t.Errorf("expected memory_1.md in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}
}

func TestMainListWithMemories(t *testing.T) {
	if os.Getenv("MAIN_LIST_WITH") == "1" {
		os.Args = []string{"sick-memory", "list"}
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

	writeMemoryFile(t, memDir, "memory_1.md", "content")

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainListWithMemories$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_LIST_WITH=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainListWithMemories subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memories in") {
		t.Errorf("expected 'Memories in' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "memory_1.md") {
		t.Errorf("expected memory_1.md in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Total memories: 1") {
		t.Errorf("expected total memories count, got:\n%s", got)
	}
}

func TestListMain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

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
	os.Args = []string{"sick-memory", "list", "--json"}
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

	ids, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 memory ID, got %d", len(ids))
	}
	if ids[0] != "memory_1.md" {
		t.Errorf("expected memory_1.md, got %v", ids[0])
	}
}
