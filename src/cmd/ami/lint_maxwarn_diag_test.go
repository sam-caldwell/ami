package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_JSON_MaxWarn_EmitsExceededRecord(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "maxwarn_diag")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\n// TODO: a\n// TODO: b\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    setLintOptions(LintOptions{MaxWarn: 1})
    defer setLintOptions(LintOptions{MaxWarn: -1})
    var out bytes.Buffer
    if err := runLint(&out, dir, true, false, false); err == nil { t.Fatalf("expected error from maxwarn exceeded") }
    dec := json.NewDecoder(&out)
    var sawExceeded bool
    var sawSummary bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "E_MAX_WARN_EXCEEDED" { sawExceeded = true }
        if m["code"] == "SUMMARY" { sawSummary = true }
    }
    if !sawExceeded || !sawSummary { t.Fatalf("expected E_MAX_WARN_EXCEEDED and SUMMARY; out=%s", out.String()) }
}

