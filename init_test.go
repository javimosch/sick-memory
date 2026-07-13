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

func TestHandleInitJSON(t *testing.T) {
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
	handleInit(cfg)
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
	if data["status"] != "initialized" {
		t.Errorf("status = %v, want %q", data["status"], "initialized")
	}
	if data["path"] != dir {
		t.Errorf("path = %v, want %q", data["path"], dir)
	}

	if _, err := os.Stat(filepath.Join(dir, "MEMORY.md")); err != nil {
		t.Errorf("expected MEMORY.md index file: %v", err)
	}
}

func TestHandleInitTextOutput(t *testing.T) {
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
	handleInit(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system initialized at") {
		t.Errorf("expected initialization message, got %q", got)
	}
	if !strings.Contains(got, dir) {
		t.Errorf("expected memory directory path, got %q", got)
	}
}

func TestHandleInitDirectoryCreationError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		file := filepath.Join(t.TempDir(), "existing-file")
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		handleInit(&Config{MemoryDir: file})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleInitDirectoryCreationError$", "-test.v")
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

func TestHandleInitIndexFileError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "MEMORY.md"), 0755); err != nil {
			t.Fatalf("failed to create blocking directory: %v", err)
		}
		handleInit(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleInitIndexFileError$", "-test.v")
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

func TestMainInitJSON(t *testing.T) {
	if os.Getenv("MAIN_INIT_JSON") == "1" {
		os.Args = []string{"sick-memory", "init", "--json", "--memory-dir", t.TempDir()}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainInitJSON$")
	cmd.Env = append(os.Environ(), "MAIN_INIT_JSON=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainInitJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"status":"initialized"`) {
		t.Errorf("expected status initialized in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}
}

func TestMainInitText(t *testing.T) {
	if os.Getenv("MAIN_INIT_TEXT") == "1" {
		dir := os.Getenv("INIT_DIR")
		os.Args = []string{"sick-memory", "init", "--memory-dir", dir}
		main()
		return
	}

	home := t.TempDir()
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainInitText$")
	cmd.Env = append(os.Environ(), "MAIN_INIT_TEXT=1", "HOME="+home, "INIT_DIR="+dir)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainInitText subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system initialized at") {
		t.Errorf("expected initialization message, got:\n%s", got)
	}
	if !strings.Contains(got, dir) {
		t.Errorf("expected memory directory path, got:\n%s", got)
	}

	if _, err := os.Stat(filepath.Join(dir, "MEMORY.md")); err != nil {
		t.Errorf("expected MEMORY.md index file: %v", err)
	}
}

func TestInitMain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()

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
	os.Args = []string{"sick-memory", "init", "--json", "--memory-dir", dir}

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

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "initialized" {
		t.Errorf("status = %v, want %q", data["status"], "initialized")
	}
	if data["path"] != dir {
		t.Errorf("path = %v, want %q", data["path"], dir)
	}

	if _, err := os.Stat(filepath.Join(dir, "MEMORY.md")); err != nil {
		t.Errorf("expected MEMORY.md index file: %v", err)
	}
}

func TestInitMainTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()

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
	os.Args = []string{"sick-memory", "init", "--memory-dir", dir}
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
	if !strings.Contains(got, "Memory system initialized at") {
		t.Errorf("expected initialization message, got %q", got)
	}
	if !strings.Contains(got, dir) {
		t.Errorf("expected memory directory path, got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dir, "MEMORY.md")); err != nil {
		t.Errorf("expected MEMORY.md index file: %v", err)
	}
}

func TestInitMainGitRepo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	if err := exec.Command("git", "-C", dir, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
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
	os.Args = []string{"sick-memory", "init", "--json"}
	jsonOutput = true
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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "initialized" {
		t.Errorf("status = %v, want %q", data["status"], "initialized")
	}

	wantMemoryDir := filepath.Join(home, ".sick-memory", "projects", sanitizePath(dir), "memory")
	if data["path"] != wantMemoryDir {
		t.Errorf("path = %v, want %q", data["path"], wantMemoryDir)
	}

	if _, err := os.Stat(filepath.Join(wantMemoryDir, "MEMORY.md")); err != nil {
		t.Errorf("expected MEMORY.md index file: %v", err)
	}
}
