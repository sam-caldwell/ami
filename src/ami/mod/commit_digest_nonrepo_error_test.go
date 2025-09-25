package mod

import (
    "path/filepath"
    "testing"
)

func TestCommitDigest_NonRepoPath_Error(t *testing.T) {
    dir := t.TempDir()
    if _, err := CommitDigestForCLI(filepath.Join(dir, "not-a-repo"), "v1.0.0"); err == nil {
        t.Fatalf("expected error when opening non-repo path")
    }
}

