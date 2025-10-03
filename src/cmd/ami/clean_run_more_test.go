package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

// Exercise the error path where stat of workspace fails with a non-ENOENT error.
func TestRunClean_StatWorkspaceError(t *testing.T) {
    base := t.TempDir()
    // Create a subdir without execute permission to trigger EACCES when statting the workspace file
    blocked := filepath.Join(base, "blocked")
    if err := os.MkdirAll(blocked, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := filepath.Join(blocked, "ami.workspace")
    if err := os.WriteFile(ws, []byte("{}"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.Chmod(blocked, 0o000); err != nil { t.Fatalf("chmod: %v", err) }
    defer os.Chmod(blocked, 0o755)

    var buf bytes.Buffer
    err := runClean(&buf, blocked, true)
    if err == nil { t.Fatalf("expected error from runClean when stat fails") }
    if exit.UnwrapCode(err) != exit.IO { t.Fatalf("expected exit.IO, got %v", exit.UnwrapCode(err)) }
}

