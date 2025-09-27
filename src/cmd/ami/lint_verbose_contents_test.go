package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Validate lint.ndjson is valid NDJSON with a final SUMMARY and at least one record.
func TestRunLint_Verbose_DebugFile_ContentsAreNDJSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "verbose_ndjson")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // Add a small source that will generate at least one warning (TODO)
    srcDir := filepath.Join(dir, ws.Packages[0].Package.Root)
    _ = os.MkdirAll(filepath.Join(dir, srcDir), 0o755)
    _ = os.MkdirAll(filepath.Join(dir, "src"), 0o755)
    _ = os.WriteFile(filepath.Join(dir, "src", "main.ami"), []byte("package x\n// TODO: check\n"), 0o644)
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var out bytes.Buffer
    if err := runLint(&out, dir, true, true, false); err != nil { /* warnings allowed */ }
    p := filepath.Join(dir, "build", "debug", "lint.ndjson")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read debug: %v", err) }
    lines := bytes.Split(bytes.TrimSpace(b), []byte("\n"))
    if len(lines) == 0 { t.Fatalf("expected at least one NDJSON line") }
    var last map[string]any
    if jerr := json.Unmarshal(lines[len(lines)-1], &last); jerr != nil { t.Fatalf("invalid JSON in last line: %v", jerr) }
    if last["code"] != "SUMMARY" { t.Fatalf("expected final SUMMARY record, got: %v", last) }
}

