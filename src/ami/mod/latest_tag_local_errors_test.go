package mod

import (
    "os"
    "path/filepath"
    "testing"

    git "github.com/go-git/go-git/v5"
)

func TestLatestTagLocal_NoTags_Error(t *testing.T) {
    dir := t.TempDir()
    if _, err := git.PlainInit(dir, false); err != nil { t.Fatalf("init: %v", err) }
    if _, err := latestTagLocal(dir); err == nil {
        t.Fatalf("expected error when no tags exist")
    }
    // Also ensure non-repo path errors
    if _, err := latestTagLocal(filepath.Join(dir, "missing")); err == nil {
        t.Fatalf("expected error when repo path invalid")
    }
    // And when path is a file
    f := filepath.Join(dir, "file")
    _ = os.WriteFile(f, []byte("x"), 0o644)
    if _, err := latestTagLocal(f); err == nil {
        t.Fatalf("expected error when path is file")
    }
}

