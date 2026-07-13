package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestHandleVersion(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	handleVersion()
	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	want := "sick-memory version " + Version + "\n"
	if got := string(out); got != want {
		t.Errorf("handleVersion output = %q, want %q", got, want)
	}
}

func TestMainVersion(t *testing.T) {
	if os.Getenv("MAIN_VERSION") == "1" {
		os.Args = []string{"sick-memory", "version"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainVersion$", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_VERSION=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainVersion subprocess failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "sick-memory version "+Version) {
		t.Errorf("expected version output, got:\n%s", out)
	}
}

func TestMainLongVersion(t *testing.T) {
	if os.Getenv("MAIN_LONG_VERSION") == "1" {
		os.Args = []string{"sick-memory", "--version"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainLongVersion$", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_LONG_VERSION=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainLongVersion subprocess failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "sick-memory version "+Version) {
		t.Errorf("expected version output, got:\n%s", out)
	}
}

func TestMainShortVersion(t *testing.T) {
	if os.Getenv("MAIN_SHORT_VERSION") == "1" {
		os.Args = []string{"sick-memory", "-v"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainShortVersion$", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_SHORT_VERSION=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainShortVersion subprocess failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "sick-memory version "+Version) {
		t.Errorf("expected version output, got:\n%s", out)
	}
}

func TestVersionMain(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"sick-memory", "version"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	main()
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	want := "sick-memory version " + Version + "\n"
	if got := string(out); got != want {
		t.Errorf("main() output = %q, want %q", got, want)
	}
}

func TestLongVersionMain(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"sick-memory", "--version"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	main()
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	want := "sick-memory version " + Version + "\n"
	if got := string(out); got != want {
		t.Errorf("main() output = %q, want %q", got, want)
	}
}

func TestShortVersionMain(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"sick-memory", "-v"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	main()
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	want := "sick-memory version " + Version + "\n"
	if got := string(out); got != want {
		t.Errorf("main() output = %q, want %q", got, want)
	}
}

func TestMainVersionMemoryDir(t *testing.T) {
	if os.Getenv("MAIN_VERSION_MEMDIR") == "1" {
		os.Args = []string{"sick-memory", "version", "--memory-dir", t.TempDir()}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainVersionMemoryDir$", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_VERSION_MEMDIR=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainVersionMemoryDir subprocess failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "sick-memory version "+Version) {
		t.Errorf("expected version output, got:\n%s", out)
	}
}
