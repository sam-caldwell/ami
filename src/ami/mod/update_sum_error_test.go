package mod

import (
	"path/filepath"
	"testing"
)

func TestUpdateSum_ErrorOnBadRepoPath(t *testing.T) {
	sumPath := filepath.Join(t.TempDir(), "ami.sum")
	if err := UpdateSum(sumPath, "example/repo", "v1.0.0", filepath.Join(t.TempDir(), "no-repo"), "v1.0.0"); err == nil {
		t.Fatalf("expected error when commit digest fails")
	}
}
