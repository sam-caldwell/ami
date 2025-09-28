package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

func TestRunLint_MissingWorkspace_ReturnsUserError(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "missing_ws")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    err := runLint(os.Stdout, dir, true, false, false)
    if err == nil { t.Fatalf("expected error") }
    if exit.UnwrapCode(err) != exit.User { t.Fatalf("want USER error code, got %v", exit.UnwrapCode(err)) }
}

