package mod

import (
    "os"
    "path/filepath"
    "testing"
)

func TestFileBackend_RejectsPathOutsideWorkspace(t *testing.T) {
    // HOME for cache
    t.Setenv("HOME", t.TempDir())
    // Workspace
    ws := t.TempDir()
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\n"), 0o644); err != nil {
        t.Fatalf("write ws: %v", err)
    }
    old, _ := os.Getwd()
    defer func(){ _ = os.Chdir(old) }()
    _ = os.Chdir(ws)

    // Create a directory outside workspace
    outside := t.TempDir()
    // Absolute path
    if _, _, _, err := GetWithInfo(outside); err == nil {
        t.Fatalf("expected error for outside-of-workspace path")
    }
}

