package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure human mode includes (file:line:col) when positions exist.
func TestLint_Human_Positions_Formatting(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "human_pos_fmt")
    src := filepath.Join(dir, "src")
    if err := os.MkdirAll(src, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(src, "main.ami"), []byte("\tline\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var out bytes.Buffer
    if err := runLint(&out, dir, false, false, false); err != nil { /* warnings allowed */ }
    s := out.String()
    if !strings.Contains(s, ":1:1)") {
        t.Fatalf("expected position formatting in human output; got: %s", s)
    }
}

