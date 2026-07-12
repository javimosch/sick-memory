package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeleteMainMemoryNotFound(t *testing.T) {
	if os.Getenv("DELETE_MAIN_MEMORY_NOT_FOUND") == "1" {
		home := os.Getenv("HOME")
		repo := os.Getenv("REPO_ROOT")
		memDir := filepath.Join(home, ".sick-memory", "projects", sanitizePath(repo), "memory")
		if err := os.MkdirAll(memDir, 0755); err != nil {
			t.Fatalf("failed to create memory dir: %v", err)
		}
		os.Args = []string{"sick-memory", "delete", "999"}
		main()
		return
	}

	home := t.TempDir()
	repo := t.TempDir()
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestDeleteMainMemoryNotFound$", "-test.v")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "DELETE_MAIN_MEMORY_NOT_FOUND=1", "HOME="+home, "REPO_ROOT="+repo)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 92 {
		t.Fatalf("expected exit code 92, got %v", err)
	}

	if !strings.Contains(string(out), "Memory with ID 999 not found") {
		t.Errorf("expected memory not found message, got:\n%s", out)
	}
}
