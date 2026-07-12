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

func TestMainNoArgs(t *testing.T) {
	if os.Getenv("MAIN_NO_ARGS") == "1" {
		os.Args = []string{"sick-memory"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainNoArgs", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_NO_ARGS=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("TestMainNoArgs expected exit error, got nil")
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 85 {
		t.Fatalf("TestMainNoArgs expected exit code 85, got %v", err)
	}

	if !strings.Contains(string(out), "sick-memory - File-based memory system for AI coding agents") {
		t.Errorf("expected help output, got:\n%s", out)
	}
}

func TestMainUnknownCommand(t *testing.T) {
	if os.Getenv("MAIN_UNKNOWN") == "1" {
		os.Args = []string{"sick-memory", "unknown"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainUnknownCommand", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_UNKNOWN=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("TestMainUnknownCommand expected exit error, got nil")
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 85 {
		t.Fatalf("TestMainUnknownCommand expected exit code 85, got %v", err)
	}

	if !strings.Contains(string(out), "Unknown command: unknown") {
		t.Errorf("expected unknown command error, got:\n%s", out)
	}
	if !strings.Contains(string(out), "sick-memory - File-based memory system for AI coding agents") {
		t.Errorf("expected help output, got:\n%s", out)
	}
}

func TestMainLongHelp(t *testing.T) {
	if os.Getenv("MAIN_LONG_HELP") == "1" {
		os.Args = []string{"sick-memory", "--help"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainLongHelp", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_LONG_HELP=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainLongHelp subprocess failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "sick-memory - File-based memory system for AI coding agents") {
		t.Errorf("expected help output, got:\n%s", out)
	}
}
