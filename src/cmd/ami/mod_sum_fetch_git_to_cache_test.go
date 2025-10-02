package main

import (
    "testing"
)

func Test_fetchGitToCache_relativeFileGit(t *testing.T) {
    if err := fetchGitToCache("file+git://relative/path", "v1.0.0", t.TempDir()); err == nil {
        t.Fatalf("expected error for non-absolute file+git source")
    }
}

