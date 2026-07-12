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

func TestConfigMainTextGitRepo(t *testing.T) {
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
	os.Args = []string{"sick-memory", "config"}
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
	want := []string{
		"Sick-Memory Configuration:",
		"Project Root: " + dir,
		"Storage Mode: Centralized (git-based scoping)",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("config text output missing %q, got:\n%s", w, got)
		}
	}
}

func TestConfigMainJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
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
	os.Args = []string{"sick-memory", "config", "--json"}
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
	if data["memory_directory"] != ".sick-memory" {
		t.Errorf("expected memory_directory '.sick-memory', got %v", data["memory_directory"])
	}
	if data["project_root"] != "" {
		t.Errorf("expected empty project_root, got %v", data["project_root"])
	}
	if data["global_directory"] != filepath.Join(home, ".sick-memory") {
		t.Errorf("expected global_directory %q, got %v", filepath.Join(home, ".sick-memory"), data["global_directory"])
	}

	globalConfig, ok := data["global_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected global_config object, got %T", data["global_config"])
	}
	if globalConfig["default_memory_type"] != "user" {
		t.Errorf("expected default_memory_type 'user', got %v", globalConfig["default_memory_type"])
	}
	if globalConfig["auto_index"] != true {
		t.Errorf("expected auto_index true, got %v", globalConfig["auto_index"])
	}
}

func TestConfigMainText(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
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
	os.Args = []string{"sick-memory", "config"}
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
	want := []string{
		"Sick-Memory Configuration:",
		"Global Directory: " + filepath.Join(home, ".sick-memory"),
		"Memory Directory: .sick-memory",
		"Project Root: Not in a git repository",
		"Storage Mode: Local (fallback)",
		"Global Configuration:",
		"Default Memory Type: user",
		"Max Memory Size: 1048576 bytes",
		"Auto Index: true",
		"Configuration File:",
		filepath.Join(home, ".sick-memory", "config.json"),
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("config text output missing %q, got:\n%s", w, got)
		}
	}
}
