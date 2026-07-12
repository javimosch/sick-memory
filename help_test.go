package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestPrintHelp(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	printHelp()
	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	want := []string{
		"sick-memory - File-based memory system for AI coding agents",
		"USAGE:",
		"COMMANDS:",
		"init",
		"remember",
		"recall",
		"OPTIONS:",
		"--json",
		"AGENT BRIDGES:",
		"EXIT CODES:",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("help output missing %q, got:\n%s", w, got)
		}
	}
}

func TestMainHelp(t *testing.T) {
	if os.Getenv("MAIN_HELP") == "1" {
		os.Args = []string{"sick-memory", "help"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelp", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_HELP=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainHelp subprocess failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "sick-memory - File-based memory system for AI coding agents") {
		t.Errorf("expected help output, got:\n%s", out)
	}
}

func TestMainHelpFlag(t *testing.T) {
	if os.Getenv("MAIN_HELP_FLAG") == "1" {
		os.Args = []string{"sick-memory", "--help"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelpFlag", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_HELP_FLAG=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainHelpFlag subprocess failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "sick-memory - File-based memory system for AI coding agents") {
		t.Errorf("expected help output, got:\n%s", out)
	}
}
