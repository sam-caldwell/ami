package mod

import (
	"errors"
	"testing"
)

func TestLatestTag_NetworkError(t *testing.T) {
	_, err := latestTag("ssh://git@invalid.invalid/org/repo.git")
	if err == nil || !errors.Is(err, ErrNetwork) {
		t.Fatalf("expected ErrNetwork; got %v", err)
	}
}

func TestResolveConstraint_NetworkError(t *testing.T) {
	_, err := resolveConstraint("ssh://git@invalid.invalid/org/repo.git", "^v1.0.0")
	if err == nil || !errors.Is(err, ErrNetwork) {
		t.Fatalf("expected ErrNetwork; got %v", err)
	}
}
