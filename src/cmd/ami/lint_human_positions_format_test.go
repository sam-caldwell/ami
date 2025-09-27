package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_SourceFormat_Positions(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "format_positions")
    src := filepath.Join(dir, "src")
    if err := os.MkdirAll(src, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // A tab at start of line and trailing spaces
    content := []byte("\tline\nwithspace   \n")
    if err := os.WriteFile(filepath.Join(src, "main.ami"), content, 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { /* ok */ }
    dec := json.NewDecoder(&buf)
    var sawTab, sawTrailing bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "W_FORMAT_TAB_INDENT" { sawTab = true }
        if m["code"] == "W_FORMAT_TRAILING_WS" { sawTrailing = true }
    }
    if !sawTab || !sawTrailing { t.Fatalf("expected format position diagnostics; out=%s", buf.String()) }
}

