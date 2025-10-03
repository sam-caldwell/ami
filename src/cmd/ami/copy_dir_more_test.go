package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_copyDir_CopiesNestedAndSkipsVCS(t *testing.T) {
    src := t.TempDir()
    dst := t.TempDir()
    // nested files and VCS dirs
    if err := os.MkdirAll(filepath.Join(src, "sub", "dir"), 0o755); err != nil { t.Fatal(err) }
    if err := os.MkdirAll(filepath.Join(src, ".svn"), 0o755); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(src, "sub", "dir", "f.txt"), []byte("data"), 0o644); err != nil { t.Fatal(err) }
    if err := copyDir(src, dst); err != nil { t.Fatal(err) }
    if _, err := os.Stat(filepath.Join(dst, ".svn")); !os.IsNotExist(err) { t.Fatal(".svn should be skipped") }
    if b, err := os.ReadFile(filepath.Join(dst, "sub", "dir", "f.txt")); err != nil || string(b) != "data" {
        t.Fatalf("expected copied file content; got %q, err=%v", string(b), err)
    }
}

