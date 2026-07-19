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
		`^name\s*=\s*"([^"]+)"`:          "sick-memory",
		`^license\s*=\s*"([^"]+)"`:       "MIT",
		`^readme\s*=\s*"([^"]+)"`:        "README.md",
		`^rust-version\s*=\s*"([^"]+)"`:  "1.56",
		`^homepage\s*=\s*"([^"]+)"`:      "https://github.com/javimosch/sick-memory",
		`^repository\s*=\s*"([^"]+)"`:    "https://github.com/javimosch/sick-memory",
		`^documentation\s*=\s*"([^"]+)"`: "https://github.com/javimosch/sick-memory#readme",
		`^description\s*=\s*"([^"]+)"`:   "File-based memory system for AI coding agents",
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

func TestCargoKeywords(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}
	content := string(data)

	re := regexp.MustCompile(`^keywords\s*=\s*\[(.*?)\]`)
	for _, line := range regexp.MustCompile(`\r?\n`).Split(content, -1) {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		vals := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(matches[1], -1)
		want := []string{"memory", "agent", "cli", "git", "worktree"}
		if len(vals) != len(want) {
			t.Errorf("expected %d keywords, got %d", len(want), len(vals))
			return
		}
		for i, w := range want {
			if vals[i][1] != w {
				t.Errorf("keyword %d = %q, want %q", i, vals[i][1], w)
			}
		}
		return
	}
	t.Error("keywords field not found in Cargo.toml")
}

func TestCargoExcludes(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}
	content := string(data)

	re := regexp.MustCompile(`exclude\s*=\s*\[([\s\S]*?)\]`)
	matches := re.FindStringSubmatch(content)
	if matches == nil {
		t.Fatal("exclude field not found in Cargo.toml")
	}

	vals := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(matches[1], -1)
	got := make([]string, 0, len(vals))
	for _, v := range vals {
		got = append(got, v[1])
	}

	want := []string{
		".github",
		".claude",
		".copilot",
		".devin",
		".opencode",
		".am-summary",
		"*.go",
		"go.mod",
		"go.work",
		"go.work.sum",
		"build.sh",
		"*_test.go",
		"docs",
		".gitignore",
		"AGENTS.md",
	}

	gotSet := make(map[string]bool, len(got))
	for _, g := range got {
		gotSet[g] = true
	}

	for _, w := range want {
		if !gotSet[w] {
			t.Errorf("expected %q in exclude list", w)
		}
	}
}

func TestCargoCategories(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}
	content := string(data)

	re := regexp.MustCompile(`^categories\s*=\s*\[(.*?)\]`)
	for _, line := range regexp.MustCompile(`\r?\n`).Split(content, -1) {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		vals := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(matches[1], -1)
		want := []string{"command-line-utilities", "development-tools"}
		if len(vals) != len(want) {
			t.Errorf("expected %d categories, got %d", len(want), len(vals))
			return
		}
		for i, w := range want {
			if vals[i][1] != w {
				t.Errorf("category %d = %q, want %q", i, vals[i][1], w)
			}
		}
		return
	}
	t.Error("categories field not found in Cargo.toml")
}

func TestCargoExcludesNoDuplicates(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}
	content := string(data)

	re := regexp.MustCompile(`exclude\s*=\s*\[([\s\S]*?)\]`)
	matches := re.FindStringSubmatch(content)
	if matches == nil {
		t.Fatal("exclude field not found in Cargo.toml")
	}

	vals := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(matches[1], -1)
	seen := make(map[string]bool, len(vals))
	for _, v := range vals {
		if seen[v[1]] {
			t.Errorf("duplicate exclude entry %q in Cargo.toml", v[1])
		}
		seen[v[1]] = true
	}
}

func TestCargoAuthors(t *testing.T) {
	data, err := os.ReadFile("Cargo.toml")
	if err != nil {
		t.Fatalf("failed to read Cargo.toml: %v", err)
	}
	content := string(data)

	re := regexp.MustCompile(`^authors\s*=\s*\[(.*?)\]`)
	for _, line := range regexp.MustCompile(`\r?\n`).Split(content, -1) {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		vals := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(matches[1], -1)
		if len(vals) == 0 {
			t.Error("authors field is empty")
			return
		}
		for _, v := range vals {
			if v[1] == "" {
				t.Errorf("empty author entry in authors field")
			}
		}
		return
	}
	t.Error("authors field not found in Cargo.toml")
}
