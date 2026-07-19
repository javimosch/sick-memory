package main

import (
	"os"
	"strings"
	"testing"
)

func TestCargoVersionMatchesGoVersion(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "version") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		cargoVersion := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		if cargoVersion != Version {
			t.Errorf("Cargo.toml version %q does not match Go Version %q", cargoVersion, Version)
		}
		return
	}

	t.Error("version field not found in Cargo.toml")
}
