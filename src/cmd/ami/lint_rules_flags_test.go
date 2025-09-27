package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_RulesFilter_And_MaxWarn_Failfast_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "rules_flags")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Underscore ident and TODO to generate multiple warnings
    content := "package x\nfunc F(){ var bad_name; // TODO: fix }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    // Apply options: filter to only W_IDENT_*, set maxWarn=0 to force error, and failfast to non-zero on warnings
    setLintOptions(LintOptions{Rules: []string{"W_IDENT_*"}, MaxWarn: 0, FailFast: true, CompatCodes: true})
    defer setLintOptions(LintOptions{MaxWarn: -1})

    var buf bytes.Buffer
    err := runLint(&buf, dir, true, false, false)
    if err == nil { t.Fatalf("expected error due to failfast/max-warn; out=%s", buf.String()) }
    dec := json.NewDecoder(&buf)
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "W_TODO" { t.Fatalf("rules filter failed; saw W_TODO: %s", buf.String()) }
        if m["code"] == "SUMMARY" {
            data := m["data"].(map[string]any)
            if data["errors"].(float64) < 1 {
                t.Fatalf("expected non-zero errors in summary due to max-warn/failfast; %s", buf.String())
            }
        }
    }
}

