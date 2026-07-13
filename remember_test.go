package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHandleRememberJSON(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: false},
	}

	os.Args = []string{"cmd", "remember", "test memory"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_"+id+".md")); err != nil {
		t.Errorf("expected memory file to exist: %v", err)
	}
}

func TestHandleRememberTextOutput(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: false},
	}

	os.Args = []string{"cmd", "remember", "test memory"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	got := string(out)
	if !strings.HasPrefix(got, "Memory saved with ID:") {
		t.Errorf("expected saved message, got %q", got)
	}
}

func TestHandleRememberAutoIndex(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: true},
	}

	os.Args = []string{"cmd", "remember", "auto index test"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}

	if _, err := os.Stat(filepath.Join(dir, "search_index.json")); err != nil {
		t.Errorf("expected search index to be written: %v", err)
	}
}

func TestHandleRememberMissingContentNoInteractive(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		noInteractive = true
		os.Args = []string{"cmd", "remember"}
		handleRemember(&Config{MemoryDir: t.TempDir()})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleRememberMissingContentNoInteractive", "-test.v")
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

func TestHandleRememberMissingContentInteractive(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		noInteractive = false
		os.Args = []string{"cmd", "remember"}
		handleRemember(&Config{MemoryDir: t.TempDir()})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleRememberMissingContentInteractive", "-test.v")
	cmd.Env = append(os.Environ(), "EXIT_TEST=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
	if exitErr.ExitCode() != 110 {
		t.Errorf("expected exit code 110, got %d", exitErr.ExitCode())
	}
}

func TestHandleRememberWriteError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		jsonOutput = true
		dir := t.TempDir()
		notDir := filepath.Join(dir, "notdir")
		if err := os.WriteFile(notDir, []byte(""), 0644); err != nil {
			t.Fatalf("failed to create notdir file: %v", err)
		}
		os.Args = []string{"cmd", "remember", "test memory"}
		handleRemember(&Config{MemoryDir: notDir, GlobalConfig: GlobalConfig{AutoIndex: false}})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleRememberWriteError", "-test.v")
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

func TestHandleRememberSkipsGlobalFlags(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: false},
	}

	os.Args = []string{"cmd", "remember", "--json", "--memory-dir", "/tmp", "flag test memory"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}

	content, err := os.ReadFile(filepath.Join(dir, "memory_"+id+".md"))
	if err != nil {
		t.Fatalf("failed to read memory file: %v", err)
	}
	if !strings.Contains(string(content), "flag test memory") {
		t.Errorf("expected memory content to contain %q, got %q", "flag test memory", string(content))
	}
}

func TestMainRememberJSON(t *testing.T) {
	if os.Getenv("MAIN_REMEMBER_JSON") == "1" {
		os.Args = []string{"sick-memory", "remember", "--json", "main remember test"}
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

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainRememberJSON$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_REMEMBER_JSON=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainRememberJSON subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"status":"remembered"`) {
		t.Errorf("expected remembered status in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}

	files, err := filepath.Glob(filepath.Join(memDir, "memory_*.md"))
	if err != nil {
		t.Fatalf("failed to glob memory files: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 memory file, got %d", len(files))
	}
}

func TestMainRememberText(t *testing.T) {
	if os.Getenv("MAIN_REMEMBER_TEXT") == "1" {
		os.Args = []string{"sick-memory", "remember", "main remember text test"}
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

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainRememberText$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_REMEMBER_TEXT=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainRememberText subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, "Memory saved with ID:") {
		t.Errorf("expected saved message, got:\n%s", got)
	}

	files, err := filepath.Glob(filepath.Join(memDir, "memory_*.md"))
	if err != nil {
		t.Fatalf("failed to glob memory files: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 memory file, got %d", len(files))
	}
}

func TestMainKeepAlias(t *testing.T) {
	if os.Getenv("MAIN_KEEP_ALIAS") == "1" {
		os.Args = []string{"sick-memory", "keep", "--json", "main keep test"}
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

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainKeepAlias$")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "MAIN_KEEP_ALIAS=1", "HOME="+home)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("TestMainKeepAlias subprocess failed: %v\n%s", err, out)
	}

	got := string(out)
	if !strings.Contains(got, `"status":"remembered"`) {
		t.Errorf("expected remembered status in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"version":"1.0"`) {
		t.Errorf("expected version 1.0 in output, got:\n%s", got)
	}

	files, err := filepath.Glob(filepath.Join(memDir, "memory_*.md"))
	if err != nil {
		t.Fatalf("failed to glob memory files: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 memory file, got %d", len(files))
	}
}

func TestRememberMain(t *testing.T) {
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
	os.Args = []string{"sick-memory", "remember", "--json", "main remember test"}
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}
	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}
	if _, err := os.ReadFile(filepath.Join(dir, "memory_"+id+".md")); err != nil {
		t.Errorf("expected memory file to exist: %v", err)
	}
}

func TestHandleRememberCreatedTimestamp(t *testing.T) {
	oldJSON := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = oldJSON })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	dir := t.TempDir()
	cfg := &Config{
		MemoryDir:    dir,
		GlobalConfig: GlobalConfig{AutoIndex: false},
	}

	os.Args = []string{"cmd", "remember", "timestamp test memory"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	handleRemember(cfg)
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
	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}

	content, err := os.ReadFile(filepath.Join(dir, "memory_"+id+".md"))
	if err != nil {
		t.Fatalf("failed to read memory file: %v", err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "created:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "created:"))
			if _, err := time.Parse(time.RFC3339, value); err != nil {
				t.Errorf("created timestamp %q is not RFC3339: %v", value, err)
			}
			return
		}
	}
	t.Errorf("memory file does not contain a created frontmatter field")
}

func TestMainRememberMemoryDir(t *testing.T) {
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
	os.Args = []string{"sick-memory", "remember", "--json", "--memory-dir", dir, "main remember memory dir test"}
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}
	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}
	if _, err := os.ReadFile(filepath.Join(dir, "memory_"+id+".md")); err != nil {
		t.Errorf("expected memory file to exist: %v", err)
	}
}

func TestMainRememberMemoryDirTextOutput(t *testing.T) {
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
	os.Args = []string{"sick-memory", "remember", "--memory-dir", dir, "main remember memory dir text test"}
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
	if !strings.HasPrefix(got, "Memory saved with ID:") {
		t.Errorf("expected saved message, got %q", got)
	}

	files, err := filepath.Glob(filepath.Join(dir, "memory_*.md"))
	if err != nil {
		t.Fatalf("failed to glob memory files: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 memory file, got %d", len(files))
	}
}

func TestKeepMainMemoryDir(t *testing.T) {
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
	os.Args = []string{"sick-memory", "keep", "--json", "--memory-dir", dir, "keep main memory dir test"}
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
	if data["status"] != "remembered" {
		t.Errorf("status = %v, want %q", data["status"], "remembered")
	}
	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got %v", data["id"])
	}
	if _, err := os.ReadFile(filepath.Join(dir, "memory_"+id+".md")); err != nil {
		t.Errorf("expected memory file to exist: %v", err)
	}
}
