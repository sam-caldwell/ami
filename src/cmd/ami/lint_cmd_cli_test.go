package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestNewLintCmd_Execute_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_lint")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    _ = os.Chdir(dir)
    c := newLintCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    if out.Len() == 0 { t.Fatalf("expected json output") }
}

