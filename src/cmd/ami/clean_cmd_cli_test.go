package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

func TestNewCleanCmd_Execute_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_clean")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Chdir into sandbox
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    _ = os.Chdir(dir)
    c := newCleanCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    if out.Len() == 0 { t.Fatalf("expected json output") }
}

