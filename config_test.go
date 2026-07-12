package main

import (
	"os"
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
