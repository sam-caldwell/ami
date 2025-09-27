package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// When workspace options include "strict", warnings are promoted to errors even if CLI strict=false.
func TestLint_Strict_FromWorkspace_Options(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "ws_strict")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\n// TODO: warn\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    // Set strict in workspace options
    ws.Toolchain.Linter.Options = []string{"strict"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var out bytes.Buffer
    // CLI strict=false, but workspace strict should elevate
    if err := runLint(&out, dir, true, false, false); err == nil {
        t.Fatalf("expected error due to workspace strict; out=%s", out.String())
    }
}

