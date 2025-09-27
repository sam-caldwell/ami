package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Suppress multiple codes at a path prefix.
func TestLint_Suppress_MultipleCodes(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "suppress_multi")
    root := filepath.Join(dir, "src")
    if err := os.MkdirAll(root, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(root, "main.ami"), []byte("// TODO: here\nUNKNOWN_IDENT\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if ws.Toolchain.Linter.Suppress == nil { ws.Toolchain.Linter.Suppress = map[string][]string{} }
    ws.Toolchain.Linter.Suppress["./src"] = []string{"W_TODO", "W_UNKNOWN_IDENT"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { /* allow */ }
    dec := json.NewDecoder(&buf)
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "W_TODO" || m["code"] == "W_UNKNOWN_IDENT" {
            t.Fatalf("expected both codes suppressed; out=%s", buf.String())
        }
    }
}

