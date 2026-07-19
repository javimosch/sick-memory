package main

import (
	"os"
	"regexp"
	"testing"
)

func TestCargoPackageName(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}

	re := regexp.MustCompile(`^name\s*=\s*"([^"]+)"`)
	for _, line := range regexp.MustCompile(`\r?\n`).Split(string(data), -1) {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		if matches[1] != "sick-memory" {
			t.Errorf("Cargo.toml name %q does not match expected %q", matches[1], "sick-memory")
		}
		return
	}

	t.Error("name field not found in Cargo.toml")
}

func TestCargoVersionMatchesGoVersion(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}

	re := regexp.MustCompile(`^version\s*=\s*"([^"]+)"`)
	for _, line := range regexp.MustCompile(`\r?\n`).Split(string(data), -1) {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		cargoVersion := matches[1]
		if cargoVersion != Version {
			t.Errorf("Cargo.toml version %q does not match Go Version %q", cargoVersion, Version)
		}
		return
	}

	t.Error("version field not found in Cargo.toml")
}

func TestCargoMetadata(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}
	content := string(data)

	checks := map[string]string{
		`^name\s*=\s*"([^"]+)"`:            "sick-memory",
		`^license\s*=\s*"([^"]+)"`:          "MIT",
		`^readme\s*=\s*"([^"]+)"`:          "README.md",
		`^rust-version\s*=\s*"([^"]+)"`:    "1.56",
		`^homepage\s*=\s*"([^"]+)"`:       "https://github.com/javimosch/sick-memory",
		`^repository\s*=\s*"([^"]+)"`:       "https://github.com/javimosch/sick-memory",
		`^description\s*=\s*"([^"]+)"`:      "File-based memory system for AI coding agents",
	}

	for pattern, want := range checks {
		re := regexp.MustCompile(pattern)
		for _, line := range regexp.MustCompile(`\r?\n`).Split(content, -1) {
			matches := re.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			if matches[1] != want {
				t.Errorf("Cargo.toml field %q = %q, want %q", pattern, matches[1], want)
			}
			delete(checks, pattern)
			break
		}
	}

	for pattern := range checks {
		t.Errorf("field matching %q not found in Cargo.toml", pattern)
	}
}
