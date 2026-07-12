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

func TestLoadGlobalConfigCreatesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := loadGlobalConfig()

	if cfg.DefaultMemoryType != "user" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "user")
	}
	if cfg.MaxMemorySize != 1024*1024 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 1024*1024)
	}
	if cfg.AutoIndex != true {
		t.Errorf("AutoIndex = %v, want true", cfg.AutoIndex)
	}

	configPath := filepath.Join(home, ".sick-memory", "config.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("expected config file at %s: %v", configPath, err)
	}
}

func TestLoadGlobalConfigIgnoresInvalidJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalDir := filepath.Join(home, ".sick-memory")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global dir: %v", err)
	}
	configPath := filepath.Join(globalDir, "config.json")
	if err := os.WriteFile(configPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := loadGlobalConfig()

	if cfg.DefaultMemoryType != "user" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "user")
	}
	if cfg.MaxMemorySize != 1024*1024 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 1024*1024)
	}
	if cfg.AutoIndex != true {
		t.Errorf("AutoIndex = %v, want true", cfg.AutoIndex)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("expected config file to be rewritten: %v", err)
	}
}

func TestLoadGlobalConfigIgnoresUnknownFields(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalDir := filepath.Join(home, ".sick-memory")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global dir: %v", err)
	}
	configPath := filepath.Join(globalDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{
  "default_memory_type": "project",
  "unknown_field": "should be ignored",
  "max_memory_size": 2048,
  "auto_index": false,
  "nested": {"ignored": true}
}`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := loadGlobalConfig()

	if cfg.DefaultMemoryType != "project" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "project")
	}
	if cfg.MaxMemorySize != 2048 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 2048)
	}
	if cfg.AutoIndex != false {
		t.Errorf("AutoIndex = %v, want false", cfg.AutoIndex)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("expected config file to remain: %v", err)
	}
}

func TestLoadGlobalConfigFillsMissingFieldsWithDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalDir := filepath.Join(home, ".sick-memory")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global dir: %v", err)
	}
	configPath := filepath.Join(globalDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"default_memory_type": "project"}`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := loadGlobalConfig()

	if cfg.DefaultMemoryType != "project" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "project")
	}
	if cfg.MaxMemorySize != 1024*1024 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 1024*1024)
	}
	if cfg.AutoIndex != true {
		t.Errorf("AutoIndex = %v, want true", cfg.AutoIndex)
	}
}

func TestLoadGlobalConfigLoadsExistingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalDir := filepath.Join(home, ".sick-memory")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global dir: %v", err)
	}
	configPath := filepath.Join(globalDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{
  "default_memory_type": "project",
  "max_memory_size": 2048,
  "auto_index": false
}`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := loadGlobalConfig()

	if cfg.DefaultMemoryType != "project" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "project")
	}
	if cfg.MaxMemorySize != 2048 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 2048)
	}
	if cfg.AutoIndex != false {
		t.Errorf("AutoIndex = %v, want false", cfg.AutoIndex)
	}
}

func TestLoadGlobalConfigHandlesEmptyConfigFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalDir := filepath.Join(home, ".sick-memory")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatalf("failed to create global dir: %v", err)
	}
	configPath := filepath.Join(globalDir, "config.json")
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := loadGlobalConfig()

	if cfg.DefaultMemoryType != "user" {
		t.Errorf("DefaultMemoryType = %q, want %q", cfg.DefaultMemoryType, "user")
	}
	if cfg.MaxMemorySize != 1024*1024 {
		t.Errorf("MaxMemorySize = %d, want %d", cfg.MaxMemorySize, 1024*1024)
	}
	if cfg.AutoIndex != true {
		t.Errorf("AutoIndex = %v, want true", cfg.AutoIndex)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("expected config file to be rewritten: %v", err)
	}
}

func TestGetGlobalSickMemoryDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := getGlobalSickMemoryDir()
	want := filepath.Join(home, ".sick-memory")
	if got != want {
		t.Errorf("getGlobalSickMemoryDir() = %q, want %q", got, want)
	}
}

func TestGetProjectMemoryPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	gitRoot := "/home/user/project"
	got := getProjectMemoryPath(gitRoot)
	want := filepath.Join(home, ".sick-memory", "projects", sanitizePath(gitRoot), "memory")
	if got != want {
		t.Errorf("getProjectMemoryPath(%q) = %q, want %q", gitRoot, got, want)
	}
}

func TestGetProjectMemoryPathSanitizesCharacters(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	gitRoot := "C:/Users/My Project/repo"
	got := getProjectMemoryPath(gitRoot)
	want := filepath.Join(home, ".sick-memory", "projects", "C--Users-My_Project-repo", "memory")
	if got != want {
		t.Errorf("getProjectMemoryPath(%q) = %q, want %q", gitRoot, got, want)
	}
}

func TestGetGlobalSickMemoryDirFallsBackWhenHomeMissing(t *testing.T) {
	t.Setenv("HOME", "")

	got := getGlobalSickMemoryDir()
	want := getDefaultMemoryDir()
	if got != want {
		t.Errorf("getGlobalSickMemoryDir() = %q, want %q", got, want)
	}
}

func TestFindGitRepositoryRoot(t *testing.T) {
	dir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed:", err)
	}

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	got, err := findGitRepositoryRoot()
	if err != nil {
		t.Fatalf("findGitRepositoryRoot() error = %v", err)
	}
	if got != dir {
		t.Errorf("findGitRepositoryRoot() = %q, want %q", got, dir)
	}
}

func TestFindGitRepositoryRootOutsideRepo(t *testing.T) {
	dir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	if _, err := findGitRepositoryRoot(); err == nil {
		t.Error("findGitRepositoryRoot() expected error when not in a git repository")
	}
}

func TestFindGitRepositoryRootFromNested(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested", "subdir")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(nested); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed:", err)
	}

	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	got, err := findGitRepositoryRoot()
	if err != nil {
		t.Fatalf("findGitRepositoryRoot() error = %v", err)
	}
	if got != dir {
		t.Errorf("findGitRepositoryRoot() = %q, want %q", got, dir)
	}
}

func TestHandleConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		ProjectRoot:  dir,
		GlobalConfig: loadGlobalConfig(),
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleConfig(cfg)
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
		t.Errorf("memory_directory = %v, want %q", data["memory_directory"], dir)
	}
	if data["project_root"] != dir {
		t.Errorf("project_root = %v, want %q", data["project_root"], dir)
	}

	globalConfig, ok := data["global_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected global_config object, got %T", data["global_config"])
	}
	if globalConfig["default_memory_type"] != "user" {
		t.Errorf("default_memory_type = %v, want %q", globalConfig["default_memory_type"], "user")
	}
}

func TestHandleConfigTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		ProjectRoot:  "",
		GlobalConfig: loadGlobalConfig(),
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleConfig(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Sick-Memory Configuration:") {
		t.Errorf("expected output to contain title, got %q", got)
	}
	if !strings.Contains(got, "Memory Directory:") {
		t.Errorf("expected output to contain memory directory, got %q", got)
	}
	if !strings.Contains(got, "Storage Mode: Local (fallback)") {
		t.Errorf("expected output to contain local storage mode, got %q", got)
	}
	if !strings.Contains(got, "Auto Index: true") {
		t.Errorf("expected output to contain auto index, got %q", got)
	}
}

func TestHandleConfigCentralizedTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	projectRoot := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		ProjectRoot:  projectRoot,
		GlobalConfig: loadGlobalConfig(),
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleConfig(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Sick-Memory Configuration:") {
		t.Errorf("expected output to contain title, got %q", got)
	}
	if !strings.Contains(got, "Project Root:") {
		t.Errorf("expected output to contain project root, got %q", got)
	}
	if !strings.Contains(got, "Storage Mode: Centralized (git-based scoping)") {
		t.Errorf("expected output to contain centralized storage mode, got %q", got)
	}
	if !strings.Contains(got, projectRoot) {
		t.Errorf("expected output to contain project root path, got %q", got)
	}
}

func TestMainConfigTextOutput(t *testing.T) {
	if os.Getenv("MAIN_CONFIG_TEXT") == "1" {
		os.Args = []string{"sick-memory", "config"}
		main()
		return
	}

	home := t.TempDir()
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainConfigTextOutput$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "MAIN_CONFIG_TEXT=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainConfigTextOutput subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Sick-Memory Configuration:") {
		t.Errorf("expected output to contain title, got:\n%s", got)
	}
	if !strings.Contains(got, "Memory Directory: .sick-memory") {
		t.Errorf("expected output to contain memory directory, got:\n%s", got)
	}
	if !strings.Contains(got, "Project Root: Not in a git repository") {
		t.Errorf("expected output to contain no git repository, got:\n%s", got)
	}
	if !strings.Contains(got, "Storage Mode: Local (fallback)") {
		t.Errorf("expected output to contain local storage mode, got:\n%s", got)
	}
	if !strings.Contains(got, "Auto Index: true") {
		t.Errorf("expected output to contain auto index, got:\n%s", got)
	}
}

func TestConfigMain(t *testing.T) {
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
	if !strings.Contains(got, "Sick-Memory Configuration:") {
		t.Errorf("expected output to contain title, got:\n%s", got)
	}
	if !strings.Contains(got, "Memory Directory: .sick-memory") {
		t.Errorf("expected output to contain memory directory, got:\n%s", got)
	}
	if !strings.Contains(got, "Project Root: Not in a git repository") {
		t.Errorf("expected output to contain no git repository, got:\n%s", got)
	}
	if !strings.Contains(got, "Storage Mode: Local (fallback)") {
		t.Errorf("expected output to contain local storage mode, got:\n%s", got)
	}
	if !strings.Contains(got, "Auto Index: true") {
		t.Errorf("expected output to contain auto index, got:\n%s", got)
	}
}

func TestMainConfigJSON(t *testing.T) {
	if os.Getenv("MAIN_CONFIG_JSON") == "1" {
		os.Args = []string{"sick-memory", "config", "--json"}
		main()
		return
	}

	home := t.TempDir()
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainConfigJSON$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "MAIN_CONFIG_JSON=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainConfigJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	globalDir := filepath.Join(home, ".sick-memory")
	if !strings.Contains(got, `"memory_directory":".sick-memory"`) {
		t.Errorf("expected memory directory in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"project_root":""`) {
		t.Errorf("expected empty project root in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"global_directory":"`+globalDir+`"`) {
		t.Errorf("expected global directory %q in output, got:\n%s", globalDir, got)
	}
	if !strings.Contains(got, `"default_memory_type":"user"`) {
		t.Errorf("expected default memory type in output, got:\n%s", got)
	}
}
