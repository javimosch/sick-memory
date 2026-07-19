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

func TestDeleteMainTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	writeMemoryFile(t, dir, "memory_1.md", "content")

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
	os.Args = []string{"sick-memory", "delete", "1"}
	jsonOutput = false
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

	got := string(out)
	if !strings.Contains(got, "Memory 1 deleted successfully") {
		t.Errorf("expected deleted message, got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_1.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}

func TestDeleteMainJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	writeMemoryFile(t, dir, "memory_1.md", "content")

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
	os.Args = []string{"sick-memory", "delete", "1", "--json"}
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

	var resp map[string]interface{}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "1" {
		t.Errorf("id = %v, want %q", data["id"], "1")
	}
	if data["status"] != "deleted" {
		t.Errorf("status = %v, want %q", data["status"], "deleted")
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_1.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}

func TestDeleteMainMissingArgument(t *testing.T) {
	if os.Getenv("DELETE_MAIN_MISSING") == "1" {
		dir := t.TempDir()
		t.Setenv("HOME", dir)
		os.Args = []string{"sick-memory", "delete"}
		memoryDir = ""
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestDeleteMainMissingArgument$", "-test.v")
	cmd.Env = append(os.Environ(), "DELETE_MAIN_MISSING=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 80 {
		t.Fatalf("expected exit code 80, got %v", err)
	}

	if !strings.Contains(string(out), "Memory ID required for delete") {
		t.Errorf("expected missing argument message, got:\n%s", out)
	}
}

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

func TestDeleteMainMemoryDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	writeMemoryFile(t, dir, "memory_1.md", "content")

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
	os.Args = []string{"sick-memory", "delete", "1", "--memory-dir", dir, "--json"}
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

	var resp map[string]interface{}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "1" {
		t.Errorf("id = %v, want %q", data["id"], "1")
	}
	if data["status"] != "deleted" {
		t.Errorf("status = %v, want %q", data["status"], "deleted")
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_1.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}

func TestDeleteMainMemoryDirTextOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	writeMemoryFile(t, dir, "memory_1.md", "content")

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
	os.Args = []string{"sick-memory", "delete", "1", "--memory-dir", dir}
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
	if !strings.Contains(got, "Memory 1 deleted successfully") {
		t.Errorf("expected deleted message, got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_1.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}

func TestDeleteMainReadDirError(t *testing.T) {
	if os.Getenv("DELETE_MAIN_READ_DIR") == "1" {
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		jsonOutput = true
		memoryDir = dir
		os.Args = []string{"sick-memory", "delete", "1"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestDeleteMainReadDirError$", "-test.v")
	cmd.Env = append(os.Environ(), "DELETE_MAIN_READ_DIR=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil\n%s", out)
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 92 {
		t.Fatalf("expected exit code 92, got %v", err)
	}

	if !strings.Contains(string(out), "Cannot read memory directory") {
		t.Errorf("expected read directory error message, got:\n%s", out)
	}
}

func TestDeleteMainRemoveError(t *testing.T) {
	if os.Getenv("DELETE_MAIN_REMOVE_ERROR") == "1" {
		dir := t.TempDir()
		memoryDir = dir
		jsonOutput = true
		os.Args = []string{"sick-memory", "delete", "1"}
		memDir := filepath.Join(dir, "memory_1.md")
		if err := os.MkdirAll(memDir, 0755); err != nil {
			t.Fatalf("failed to create memory directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(memDir, "keep.txt"), []byte("keep"), 0644); err != nil {
			t.Fatalf("failed to create file inside memory directory: %v", err)
		}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestDeleteMainRemoveError$", "-test.v")
	cmd.Env = append(os.Environ(), "DELETE_MAIN_REMOVE_ERROR=1", "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit error, got nil\n%s", out)
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 110 {
		t.Fatalf("expected exit code 110, got %v", err)
	}

	if !strings.Contains(string(out), "Cannot delete memory file") {
		t.Errorf("expected delete file error message, got:\n%s", out)
	}
}

func TestDeleteMainByFullMemoryID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	writeMemoryFile(t, dir, "memory_1.md", "content")

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
	os.Args = []string{"sick-memory", "delete", "memory_1", "--json"}
	jsonOutput = false
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

	var resp map[string]interface{}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "memory_1" {
		t.Errorf("id = %v, want %q", data["id"], "memory_1")
	}
	if data["status"] != "deleted" {
		t.Errorf("status = %v, want %q", data["status"], "deleted")
	}

	if _, err := os.Stat(filepath.Join(dir, "memory_1.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}

func TestDeleteMainBySuffix(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	writeMemoryFile(t, dir, "project_1.md", "content")

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
	os.Args = []string{"sick-memory", "delete", "1", "--json"}
	jsonOutput = false
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

	var resp map[string]interface{}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["id"] != "1" {
		t.Errorf("id = %v, want %q", data["id"], "1")
	}
	if data["status"] != "deleted" {
		t.Errorf("status = %v, want %q", data["status"], "deleted")
	}

	if _, err := os.Stat(filepath.Join(dir, "project_1.md")); !os.IsNotExist(err) {
		t.Errorf("expected memory file to be deleted")
	}
}
