package main

import (
	"encoding/json"
	"io"
	"os"
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
