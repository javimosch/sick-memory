package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleBridgeClaudeCode(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	os.Args = []string{"cmd", "bridge", "claude-code"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleBridge(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "bridge_created" {
		t.Errorf("status = %v, want %q", data["status"], "bridge_created")
	}
	if data["agent"] != "claude-code" {
		t.Errorf("agent = %v, want %q", data["agent"], "claude-code")
	}

	content, err := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read bridge file: %v", err)
	}
	if !strings.Contains(string(content), dir) {
		t.Errorf("expected bridge file to contain memory directory, got %q", string(content))
	}
}

func TestHandleBridgeTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	os.Args = []string{"cmd", "bridge", "opencode"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleBridge(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "OpenCode bridge created successfully") {
		t.Errorf("expected success message, got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "memory.json")); err != nil {
		t.Errorf("expected .opencode/memory.json to exist: %v", err)
	}
}

func TestHandleBridgeCopilot(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	os.Args = []string{"cmd", "bridge", "copilot"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleBridge(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "bridge_created" {
		t.Errorf("status = %v, want %q", data["status"], "bridge_created")
	}
	if data["agent"] != "copilot" {
		t.Errorf("agent = %v, want %q", data["agent"], "copilot")
	}

	content, err := os.ReadFile(filepath.Join(dir, ".copilot", "settings.json"))
	if err != nil {
		t.Fatalf("failed to read bridge file: %v", err)
	}
	if !strings.Contains(string(content), dir) {
		t.Errorf("expected bridge file to contain memory directory, got %q", string(content))
	}
}

func TestHandleBridgeCopilotTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	os.Args = []string{"cmd", "bridge", "copilot"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleBridge(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Copilot bridge created successfully") {
		t.Errorf("expected success message, got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dir, ".copilot", "settings.json")); err != nil {
		t.Errorf("expected .copilot/settings.json to exist: %v", err)
	}
}

func TestHandleBridgeClaudeCodeTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	os.Args = []string{"cmd", "bridge", "claude-code"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleBridge(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Claude Code bridge created successfully") {
		t.Errorf("expected success message, got %q", got)
	}
	if !strings.Contains(got, "Configuration file: .claude/CLAUDE.md") {
		t.Errorf("expected configuration file path, got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "CLAUDE.md")); err != nil {
		t.Errorf("expected .claude/CLAUDE.md to exist: %v", err)
	}
}

func TestHandleBridgeOpenCodeJSON(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	os.Args = []string{"cmd", "bridge", "opencode"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleBridge(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "bridge_created" {
		t.Errorf("status = %v, want %q", data["status"], "bridge_created")
	}
	if data["agent"] != "opencode" {
		t.Errorf("agent = %v, want %q", data["agent"], "opencode")
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "memory.json")); err != nil {
		t.Errorf("expected .opencode/memory.json to exist: %v", err)
	}
}

func TestHandleBridgeSkipsGlobalFlags(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	os.Args = []string{"cmd", "bridge", "--json", "--memory-dir", "/tmp", "claude-code"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleBridge(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["agent"] != "claude-code" {
		t.Errorf("agent = %v, want %q", data["agent"], "claude-code")
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "CLAUDE.md")); err != nil {
		t.Errorf("expected .claude/CLAUDE.md to exist: %v", err)
	}
}

func TestHandleBridgeMissingArgs(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		os.Args = []string{"cmd", "bridge"}
		handleBridge(&Config{MemoryDir: ""})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleBridgeMissingArgs", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 85 {
		t.Errorf("expected exit code 85, got %d", exitErr.ExitCode())
	}
}

func TestHandleBridgeUnknownAgent(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		os.Args = []string{"cmd", "bridge", "unknown"}
		handleBridge(&Config{MemoryDir: ""})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleBridgeUnknownAgent", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 85 {
		t.Errorf("expected exit code 85, got %d", exitErr.ExitCode())
	}
}

func TestGenerateClaudeCodeBridgeWriteError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".claude", "CLAUDE.md"), 0755); err != nil {
			t.Fatalf("failed to create blocking directory: %v", err)
		}
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		generateClaudeCodeBridge(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGenerateClaudeCodeBridgeWriteError", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}

func TestGenerateClaudeCodeBridgeMkdirError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".claude"), []byte("not a directory"), 0644); err != nil {
			t.Fatalf("failed to create blocking file: %v", err)
		}
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		generateClaudeCodeBridge(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGenerateClaudeCodeBridgeMkdirError", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}

func TestGenerateOpenCodeBridgeMkdirError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".opencode"), []byte("not a directory"), 0644); err != nil {
			t.Fatalf("failed to create blocking file: %v", err)
		}
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		generateOpenCodeBridge(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGenerateOpenCodeBridgeMkdirError", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}

func TestGenerateOpenCodeBridgeWriteError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".opencode", "memory.json"), 0755); err != nil {
			t.Fatalf("failed to create blocking directory: %v", err)
		}
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		generateOpenCodeBridge(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGenerateOpenCodeBridgeWriteError", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}

func TestGenerateCopilotBridgeWriteError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".copilot", "settings.json"), 0755); err != nil {
			t.Fatalf("failed to create blocking directory: %v", err)
		}
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		generateCopilotBridge(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGenerateCopilotBridgeWriteError", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}

func TestGenerateCopilotBridgeMkdirError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".copilot"), []byte("not a directory"), 0644); err != nil {
			t.Fatalf("failed to create blocking file: %v", err)
		}
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		generateCopilotBridge(&Config{MemoryDir: dir})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGenerateCopilotBridgeMkdirError", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 92 {
		t.Errorf("expected exit code 92, got %d", exitErr.ExitCode())
	}
}
