package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_JSON_CompatCodes_Off_NoField(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "compat_codes_off")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\nfunc F(){ var bad_name }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    setLintOptions(LintOptions{CompatCodes: false})
    defer setLintOptions(LintOptions{})
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { /* warnings allowed */ }
    dec := json.NewDecoder(&buf)
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "SUMMARY" { break }
        if data, ok := m["data"].(map[string]any); ok {
            if _, ok2 := data["compatCode"]; ok2 {
                t.Fatalf("compatCode should not be present when disabled; out=%s", buf.String())
            }
        }
    }
}

