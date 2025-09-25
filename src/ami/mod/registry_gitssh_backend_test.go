package mod

import (
	"errors"
	"testing"
)

// TestGitSSHBackend_NetworkError ensures network errors are surfaced with ErrNetwork sentinel.
func TestGitSSHBackend_NetworkError(t *testing.T) {
	// Invalid host is expected to fail quickly
	_, _, _, err := GetWithInfo("git+ssh://invalid.invalid/org/repo.git#v1.2.3")
	if err == nil {
		t.Fatalf("expected error for invalid host")
	}
	if !errors.Is(err, ErrNetwork) {
		t.Fatalf("expected ErrNetwork sentinel; got %v", err)
	}
}
