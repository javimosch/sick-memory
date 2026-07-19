package main

import (
	"os"
	"regexp"
	"testing"
)

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
