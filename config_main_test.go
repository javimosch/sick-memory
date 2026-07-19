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

func TestConfigMainMemoryDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

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
	os.Args = []string{"sick-memory", "config", "--json", "--memory-dir", dir}
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
	if data["memory_directory"] != dir {
		t.Errorf("expected memory_directory %q, got %v", dir, data["memory_directory"])
	}
	if data["project_root"] != "" {
		t.Errorf("expected empty project_root, got %v", data["project_root"])
	}
	if data["global_directory"] != filepath.Join(home, ".sick-memory") {
		t.Errorf("expected global_directory %q, got %v", filepath.Join(home, ".sick-memory"), data["global_directory"])
	}
}

func TestConfigMainMemoryDirTextOutput(t *testing.T) {
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
	os.Args = []string{"sick-memory", "config", "--memory-dir", dir}
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
		"Memory Directory: " + dir,
		"Project Root: Not in a git repository",
		"Storage Mode: Local (fallback)",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("config text output missing %q, got:\n%s", w, got)
		}
	}
}

func TestConfigMainJSONGitRepo(t *testing.T) {
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
	os.Args = []string{"sick-memory", "config", "--json"}
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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}

	wantMemoryDir := filepath.Join(home, ".sick-memory", "projects", sanitizePath(dir), "memory")
	if data["memory_directory"] != wantMemoryDir {
		t.Errorf("expected memory_directory %q, got %v", wantMemoryDir, data["memory_directory"])
	}
	if data["project_root"] != dir {
		t.Errorf("expected project_root %q, got %v", dir, data["project_root"])
	}
	if data["global_directory"] != filepath.Join(home, ".sick-memory") {
		t.Errorf("expected global_directory %q, got %v", filepath.Join(home, ".sick-memory"), data["global_directory"])
	}
}

func TestConfigMainMemoryDirOverridesGitRepo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	if err := exec.Command("git", "-C", dir, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	memoryDirOverride := t.TempDir()

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
	os.Args = []string{"sick-memory", "config", "--json", "--memory-dir", memoryDirOverride}
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

	if data["memory_directory"] != memoryDirOverride {
		t.Errorf("expected memory_directory %q, got %v", memoryDirOverride, data["memory_directory"])
	}
	if data["project_root"] != "" {
		t.Errorf("expected empty project_root when --memory-dir is used, got %v", data["project_root"])
	}
	if data["global_directory"] != filepath.Join(home, ".sick-memory") {
		t.Errorf("expected global_directory %q, got %v", filepath.Join(home, ".sick-memory"), data["global_directory"])
	}
}

func TestConfigMainMemoryDirOverridesGitRepoTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	if err := exec.Command("git", "-C", dir, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	memoryDirOverride := t.TempDir()

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
	os.Args = []string{"sick-memory", "config", "--memory-dir", memoryDirOverride}
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
		"Memory Directory: " + memoryDirOverride,
		"Project Root: Not in a git repository",
		"Storage Mode: Local (fallback)",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("config text output missing %q, got:\n%s", w, got)
		}
	}
}

func TestConfigMainJSONWithGlobalConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".sick-memory"), 0755); err != nil {
		t.Fatalf("failed to create global dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".sick-memory", "config.json"), []byte(`{
  "default_memory_type": "project",
  "max_memory_size": 4096,
  "auto_index": false
}`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

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
	globalConfig, ok := data["global_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected global_config object, got %T", data["global_config"])
	}
	if globalConfig["default_memory_type"] != "project" {
		t.Errorf("expected default_memory_type 'project', got %v", globalConfig["default_memory_type"])
	}
	if max, ok := globalConfig["max_memory_size"].(float64); !ok || max != 4096 {
		t.Errorf("expected max_memory_size 4096, got %v", globalConfig["max_memory_size"])
	}
	if globalConfig["auto_index"] != false {
		t.Errorf("expected auto_index false, got %v", globalConfig["auto_index"])
	}
}

func TestConfigMainShortJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

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
	os.Args = []string{"sick-memory", "config", "-j", "--memory-dir", dir}
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
	if data["memory_directory"] != dir {
		t.Errorf("expected memory_directory %q, got %v", dir, data["memory_directory"])
	}
	if data["project_root"] != "" {
		t.Errorf("expected empty project_root, got %v", data["project_root"])
	}
	if data["global_directory"] != filepath.Join(home, ".sick-memory") {
		t.Errorf("expected global_directory %q, got %v", filepath.Join(home, ".sick-memory"), data["global_directory"])
	}
}
