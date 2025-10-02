package main

import (
    "path/filepath"
    "testing"
)

func Test_listGitTags_nonRepoPath(t *testing.T) {
    // Use a temp directory which is not a git repo; expect error
    dir := t.TempDir()
    abs := filepath.Clean(dir)
    if _, err := listGitTags(abs); err == nil {
        t.Fatalf("expected error for non-repo path")
    }
}

