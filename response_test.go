package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSuccessResponse(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	successResponse(map[string]interface{}{"status": "ok"})
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

	if resp.Version != "1.0" {
		t.Errorf("version = %q, want %q", resp.Version, "1.0")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if data["status"] != "ok" {
		t.Errorf("status = %v, want %q", data["status"], "ok")
	}
}

func TestSuccessResponseWithNil(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	successResponse(nil)
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

	if resp.Version != "1.0" {
		t.Errorf("version = %q, want %q", resp.Version, "1.0")
	}
	if resp.Data != nil {
		t.Errorf("expected nil data, got %v", resp.Data)
	}
}

func TestErrorResponse(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	errorResponse(42, "test_error", "something went wrong", true)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	if resp.Error.Code != 42 {
		t.Errorf("code = %d, want 42", resp.Error.Code)
	}
	if resp.Error.Type != "test_error" {
		t.Errorf("type = %q, want %q", resp.Error.Type, "test_error")
	}
	if resp.Error.Message != "something went wrong" {
		t.Errorf("message = %q, want %q", resp.Error.Message, "something went wrong")
	}
	if !resp.Error.Recoverable {
		t.Errorf("recoverable = false, want true")
	}
}

func TestErrorResponseNonRecoverable(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	errorResponse(42, "test_error", "something went wrong", false)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	if resp.Error.Recoverable {
		t.Errorf("recoverable = true, want false")
	}
}

func TestErrorResponseSpecialCharacters(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	errorResponse(42, "test_error", "quote: \" and newline: \n and tab: \t", true)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	if resp.Error.Message != "quote: \" and newline: \n and tab: \t" {
		t.Errorf("message = %q, want %q", resp.Error.Message, "quote: \" and newline: \n and tab: \t")
	}
}

func TestErrorResponseInvalidUTF8(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	errorResponse(42, "test_error", string([]byte{0xff}), true)
	os.Stdout = old
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\n%s", err, out)
	}

	if resp.Error.Message != "\ufffd" {
		t.Errorf("message = %q, want %q", resp.Error.Message, "\ufffd")
	}
}

func TestSuccessResponseMarshalError(t *testing.T) {
	if os.Getenv("EXIT_TEST") == "1" {
		cyclic := make(map[string]interface{})
		cyclic["self"] = cyclic
		successResponse(cyclic)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestSuccessResponseMarshalError", "-test.v")
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

func TestMainErrorResponse(t *testing.T) {
	if os.Getenv("MAIN_ERROR") == "1" {
		os.Args = []string{"sick-memory", "remember", "--no-interactive"}
		main()
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainErrorResponse", "-test.v")
	cmd.Env = append(os.Environ(), "MAIN_ERROR=1", "HOME="+home)
	out, err := cmd.CombinedOutput()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got %v\n%s", err, out)
	}
	if exitErr.ExitCode() != 85 {
		t.Errorf("expected exit code 85, got %d", exitErr.ExitCode())
	}

	if !strings.Contains(string(out), "Content required for remember command") {
		t.Errorf("expected error response output, got:\n%s", out)
	}
}

func TestSuccessResponseWithArray(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	successResponse([]string{"one", "two"})
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

	data, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", resp.Data)
	}
	if len(data) != 2 {
		t.Errorf("expected 2 items, got %d", len(data))
	}
}

func TestSuccessResponseWithNumber(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	successResponse(42)
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

	data, ok := resp.Data.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", resp.Data)
	}
	if data != 42 {
		t.Errorf("data = %v, want 42", data)
	}
}

func TestSuccessResponseWithString(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	successResponse("ok")
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

	data, ok := resp.Data.(string)
	if !ok {
		t.Fatalf("expected string, got %T", resp.Data)
	}
	if data != "ok" {
		t.Errorf("data = %v, want ok", data)
	}
}

func TestSuccessResponseWithBoolean(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	successResponse(true)
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

	data, ok := resp.Data.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T", resp.Data)
	}
	if !data {
		t.Errorf("data = %v, want true", data)
	}
}
