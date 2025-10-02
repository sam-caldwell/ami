package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_copyDir_skipsVCS(t *testing.T) {
    src := t.TempDir()
    dst := t.TempDir()
    // create files and VCS dir
    _ = os.MkdirAll(filepath.Join(src, ".git"), 0o755)
    _ = os.WriteFile(filepath.Join(src, "a.txt"), []byte("ok"), 0o644)
    if err := copyDir(src, dst); err != nil { t.Fatal(err) }
    if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) { t.Fatal(".git should be skipped") }
    if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil { t.Fatal(err) }
}

