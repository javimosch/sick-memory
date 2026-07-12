package main

import (
	"os"
	"os/exec"
	"path/filepath"
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
