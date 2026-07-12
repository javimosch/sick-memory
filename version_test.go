package main

import (
	"io"
	"os"
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
