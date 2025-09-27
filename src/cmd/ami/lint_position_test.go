package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Validates that source diagnostics include positions when available.
func TestLint_SourcePositions_IncludedInJSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "positions")
    src := filepath.Join(dir, "src")
    if err := os.MkdirAll(src, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Line with sentinel at column 3
    if err := os.WriteFile(filepath.Join(src, "main.ami"), []byte("aa\nxxUNKNOWN_IDENT\n"), 0o644); err != nil {
        t.Fatalf("write: %v", err)
    }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { t.Fatalf("runLint: %v", err) }
    dec := json.NewDecoder(&buf)
    var sawPos bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_UNKNOWN_IDENT" {
            // Expect nested pos object with line and column
            if pos, ok := m["pos"].(map[string]any); ok {
                if pos["line"] == float64(2) { sawPos = true }
            }
        }
    }
    if !sawPos { t.Fatalf("expected position included for W_UNKNOWN_IDENT; out=%s", buf.String()) }
}

