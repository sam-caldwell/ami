package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestScanSourceUnknown_EmitsWarningAndSummary(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "src_scan")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    // Create a dummy .ami file with UNKNOWN_IDENT sentinel
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte("start\nUNKNOWN_IDENT\nend\n"), 0o644); err != nil {
        t.Fatalf("write: %v", err)
    }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{} // ensure non-strict for this JSON test
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
        t.Fatalf("save: %v", err)
    }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        t.Fatalf("runLint: %v", err)
    }
    // Expect 2+ JSON lines (one warning for unknown ident, plus a summary)
    dec := json.NewDecoder(&buf)
    var warns int
    var summary bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil {
            t.Fatalf("json: %v", err)
        }
        if m["code"] == "W_UNKNOWN_IDENT" {
            warns++
        }
        if m["code"] == "SUMMARY" {
            summary = true
        }
    }
    if warns < 1 || !summary {
        t.Fatalf("expected at least one W_UNKNOWN_IDENT and a SUMMARY; out=%s", buf.String())
    }
}
