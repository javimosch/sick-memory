package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleStatusUninitialized(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: filepath.Join(dir, "does-not-exist")}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
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
	if data["status"] != "uninitialized" {
		t.Errorf("expected status 'uninitialized', got %v", data["status"])
	}
}

func TestHandleStatusActive(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
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
	if data["status"] != "active" {
		t.Errorf("expected status 'active', got %v", data["status"])
	}
	if data["path"] != dir {
		t.Errorf("expected path %q, got %v", dir, data["path"])
	}
	if count, ok := data["count"].(float64); !ok || count != 0 {
		t.Errorf("expected count 0, got %v", data["count"])
	}
}

func TestHandleStatusUninitializedTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: filepath.Join(dir, "does-not-exist")}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system status: uninitialized") {
		t.Errorf("expected uninitialized status, got %q", got)
	}
	if !strings.Contains(got, "Run 'sick-memory init' to initialize.") {
		t.Errorf("expected init hint, got %q", got)
	}
}

func TestHandleStatusActiveTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system status: active") {
		t.Errorf("expected active status, got %q", got)
	}
	if !strings.Contains(got, "Memory directory:") {
		t.Errorf("expected memory directory, got %q", got)
	}
	if !strings.Contains(got, "Total memories: 0") {
		t.Errorf("expected total memories, got %q", got)
	}
}

func TestHandleStatusActiveWithMemories(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	dir := t.TempDir()
	cfg := &Config{MemoryDir: dir}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("not a memory"), 0644); err != nil {
		t.Fatalf("failed to write README file: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleStatus(cfg)
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
	if data["status"] != "active" {
		t.Errorf("expected status 'active', got %v", data["status"])
	}
	if count, ok := data["count"].(float64); !ok || count != 1 {
		t.Errorf("expected count 1, got %v", data["count"])
	}
	if data["path"] != dir {
		t.Errorf("expected path %q, got %v", dir, data["path"])
	}
}

func TestHandleStatusReadDirError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true

		dir := t.TempDir()
		memoryPath := filepath.Join(dir, "not-a-directory")
		if err := os.WriteFile(memoryPath, []byte(""), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write file: %v\n", err)
			os.Exit(1)
		}

		handleStatus(&Config{MemoryDir: memoryPath})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleStatusReadDirError", "-test.v")
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

func TestMainStatus(t *testing.T) {
	if os.Getenv("MAIN_STATUS") == "1" {
		os.Args = []string{"sick-memory", "status"}
		main()
		return
	}

	home := t.TempDir()
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainStatus$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "MAIN_STATUS=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainStatus subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system status: uninitialized") {
		t.Errorf("expected uninitialized status, got:\n%s", got)
	}
	if !strings.Contains(got, "Run 'sick-memory init' to initialize.") {
		t.Errorf("expected init hint, got:\n%s", got)
	}
}

func TestMainStatusActive(t *testing.T) {
	if os.Getenv("MAIN_STATUS_ACTIVE") == "1" {
		os.Args = []string{"sick-memory", "status"}
		main()
		return
	}

	home := t.TempDir()
	repo := t.TempDir()
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	memDir := filepath.Join(home, ".sick-memory", "projects", sanitizePath(repo), "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	writeMemoryFile(t, memDir, "memory_1.md", "content")

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainStatusActive$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_STATUS_ACTIVE=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainStatusActive subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system status: active") {
		t.Errorf("expected active status, got:\n%s", got)
	}
	if !strings.Contains(got, "Total memories: 1") {
		t.Errorf("expected total memories count, got:\n%s", got)
	}
}

func TestMainStatusJSON(t *testing.T) {
	if os.Getenv("MAIN_STATUS_JSON") == "1" {
		os.Args = []string{"sick-memory", "status", "--json"}
		main()
		return
	}

	home := t.TempDir()
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainStatusJSON$", "-test.v")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "MAIN_STATUS_JSON=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainStatusJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"status":"uninitialized"`) {
		t.Errorf("expected uninitialized status, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}
}

func TestStatusMain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	oldArgs := os.Args
	oldJSON := jsonOutput
	oldMemoryDir := memoryDir
	oldNoInteractive := noInteractive
	defer func() {
		os.Args = oldArgs
		jsonOutput = oldJSON
		memoryDir = oldMemoryDir
		noInteractive = oldNoInteractive
	}()

	os.Args = []string{"sick-memory", "status"}
	jsonOutput = false
	memoryDir = ""
	noInteractive = false

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

	got := string(out)
	if !strings.Contains(got, "Memory system status: uninitialized") {
		t.Errorf("expected uninitialized status, got:\n%s", got)
	}
	if !strings.Contains(got, "Run 'sick-memory init' to initialize.") {
		t.Errorf("expected init hint, got:\n%s", got)
	}
}

func TestStatusMainJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	oldArgs := os.Args
	oldJSON := jsonOutput
	oldMemoryDir := memoryDir
	oldNoInteractive := noInteractive
	defer func() {
		os.Args = oldArgs
		jsonOutput = oldJSON
		memoryDir = oldMemoryDir
		noInteractive = oldNoInteractive
	}()

	os.Args = []string{"sick-memory", "status", "--json"}
	jsonOutput = true
	memoryDir = dir
	noInteractive = false

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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "active" {
		t.Errorf("expected status 'active', got %v", data["status"])
	}
	if count, ok := data["count"].(float64); !ok || count != 1 {
		t.Errorf("expected count 1, got %v", data["count"])
	}
	if data["path"] != dir {
		t.Errorf("expected path %q, got %v", dir, data["path"])
	}
}

func TestStatusMainMemoryDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	oldArgs := os.Args
	oldJSON := jsonOutput
	oldMemoryDir := memoryDir
	oldNoInteractive := noInteractive
	defer func() {
		os.Args = oldArgs
		jsonOutput = oldJSON
		memoryDir = oldMemoryDir
		noInteractive = oldNoInteractive
	}()

	os.Args = []string{"sick-memory", "status", "--json", "--memory-dir", dir}
	jsonOutput = true
	memoryDir = ""
	noInteractive = false

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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "active" {
		t.Errorf("expected status 'active', got %v", data["status"])
	}
	if count, ok := data["count"].(float64); !ok || count != 1 {
		t.Errorf("expected count 1, got %v", data["count"])
	}
	if data["path"] != dir {
		t.Errorf("expected path %q, got %v", dir, data["path"])
	}
}

func TestStatusMainActiveMemoryDirTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	oldArgs := os.Args
	oldJSON := jsonOutput
	oldMemoryDir := memoryDir
	oldNoInteractive := noInteractive
	defer func() {
		os.Args = oldArgs
		jsonOutput = oldJSON
		memoryDir = oldMemoryDir
		noInteractive = oldNoInteractive
	}()

	os.Args = []string{"sick-memory", "status", "--memory-dir", dir}
	jsonOutput = false
	memoryDir = ""
	noInteractive = false

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

	got := string(out)
	if !strings.Contains(got, "Memory system status: active") {
		t.Errorf("expected active status, got %q", got)
	}
	if !strings.Contains(got, "Memory directory: "+dir) {
		t.Errorf("expected memory directory %q, got %q", dir, got)
	}
	if !strings.Contains(got, "Total memories: 1") {
		t.Errorf("expected total memories count, got %q", got)
	}
}

func TestMainStatusActiveMemoryDirJSON(t *testing.T) {
	if os.Getenv("MAIN_STATUS_ACTIVE_MEMORY_DIR_JSON") == "1" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write memory file: %v\n", err)
			os.Exit(1)
		}
		os.Args = []string{"sick-memory", "status", "--json", "--memory-dir", dir}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainStatusActiveMemoryDirJSON$")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "MAIN_STATUS_ACTIVE_MEMORY_DIR_JSON=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainStatusActiveMemoryDirJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"status":"active"`) {
		t.Errorf("expected active status in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"count":1`) {
		t.Errorf("expected count 1 in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}
}

func TestStatusMainMemoryDirUninitialized(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()

	oldArgs := os.Args
	oldJSON := jsonOutput
	oldMemoryDir := memoryDir
	oldNoInteractive := noInteractive
	defer func() {
		os.Args = oldArgs
		jsonOutput = oldJSON
		memoryDir = oldMemoryDir
		noInteractive = oldNoInteractive
	}()
	os.Args = []string{"sick-memory", "status", "--memory-dir", filepath.Join(dir, "does-not-exist")}
	jsonOutput = false
	memoryDir = ""
	noInteractive = false

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

	got := string(out)
	if !strings.Contains(got, "Memory system status: uninitialized") {
		t.Errorf("expected uninitialized status, got:\n%s", got)
	}
	if !strings.Contains(got, "Run 'sick-memory init' to initialize.") {
		t.Errorf("expected init hint, got:\n%s", got)
	}
}

func TestStatusMainMemoryDirUninitializedJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()

	oldArgs := os.Args
	oldJSON := jsonOutput
	oldMemoryDir := memoryDir
	oldNoInteractive := noInteractive
	defer func() {
		os.Args = oldArgs
		jsonOutput = oldJSON
		memoryDir = oldMemoryDir
		noInteractive = oldNoInteractive
	}()
	os.Args = []string{"sick-memory", "status", "--json", "--memory-dir", filepath.Join(dir, "does-not-exist")}
	jsonOutput = false
	memoryDir = ""
	noInteractive = false

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

	var resp SuccessResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "uninitialized" {
		t.Errorf("expected status 'uninitialized', got %v", data["status"])
	}
	if _, ok := data["count"]; ok {
		t.Errorf("did not expect count in uninitialized response")
	}
	if _, ok := data["path"]; ok {
		t.Errorf("did not expect path in uninitialized response")
	}
}

func TestMainStatusActiveMemoryDirTextOutput(t *testing.T) {
	if os.Getenv("MAIN_STATUS_ACTIVE_MEMORY_DIR_TEXT") == "1" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "memory_1.md"), []byte("content"), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write memory file: %v\n", err)
			os.Exit(1)
		}
		os.Args = []string{"sick-memory", "status", "--memory-dir", dir}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainStatusActiveMemoryDirTextOutput$")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "MAIN_STATUS_ACTIVE_MEMORY_DIR_TEXT=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("TestMainStatusActiveMemoryDirTextOutput subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memory system status: active") {
		t.Errorf("expected active status, got:\n%s", got)
	}
	if !strings.Contains(got, "Memory directory: ") {
		t.Errorf("expected memory directory line, got:\n%s", got)
	}
	if !strings.Contains(got, "Total memories: 1") {
		t.Errorf("expected total memories count, got:\n%s", got)
	}
}
