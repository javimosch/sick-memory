package main

import (
	"io"
	"os"
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
