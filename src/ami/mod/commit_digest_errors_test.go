package mod

import (
    "testing"
    "github.com/go-git/go-git/v5"
)

func TestCommitDigest_ErrorsOnMissingTag(t *testing.T) {
    dir := t.TempDir()
    if _, err := git.PlainInit(dir, false); err != nil { t.Fatalf("init: %v", err) }
    if _, err := CommitDigestForCLI(dir, "v0.0.0"); err == nil {
        t.Fatalf("expected error for missing tag")
    }
}

