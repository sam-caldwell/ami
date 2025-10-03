package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_copyDir_CopiesFilesAndSkipsVCS(t *testing.T) {
    src := t.TempDir()
    // Create files and VCS dirs
    if err := os.MkdirAll(filepath.Join(src, "a", ".git"), 0o755); err != nil { t.Fatal(err) }
    if err := os.MkdirAll(filepath.Join(src, "b", ".hg"), 0o755); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(src, "a", "f1.txt"), []byte("x"), 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(src, "b", "f2.txt"), []byte("y"), 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(src, "a", ".git", "should_skip"), []byte("z"), 0o644); err != nil { t.Fatal(err) }
    dst := t.TempDir()
    if err := copyDir(src, dst); err != nil { t.Fatalf("copyDir: %v", err) }
    // Expect files copied
    if _, err := os.Stat(filepath.Join(dst, "a", "f1.txt")); err != nil { t.Fatalf("missing f1: %v", err) }
    if _, err := os.Stat(filepath.Join(dst, "b", "f2.txt")); err != nil { t.Fatalf("missing f2: %v", err) }
    // VCS contents should be skipped
    if _, err := os.Stat(filepath.Join(dst, "a", ".git", "should_skip")); !os.IsNotExist(err) {
        t.Fatalf("expected .git contents to be skipped; err=%v", err)
    }
}

func Test_copyDir_SourceMissingReturnsError(t *testing.T) {
    dst := t.TempDir()
    if err := copyDir(filepath.Join(dst, "does_not_exist"), filepath.Join(dst, "out")); err == nil {
        t.Fatal("expected error for missing source directory")
    }
}

