package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "regexp"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Parity for explicit error rule: E_IO_PERMISSION
func TestLint_Parity_Errors_IOPermission(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "parity_io")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\npipeline P(){ ingress(); Alpha(); io.Read(); egress(); }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true})
    defer setRuleToggles(RuleToggles{})

    // JSON
    var js bytes.Buffer
    if err := runLint(&js, dir, true, false, false); err == nil { t.Fatalf("expected error; out=%s", js.String()) }
    dec := json.NewDecoder(&js)
    var errors int
    var sawCode bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "E_IO_PERMISSION" { sawCode = true }
        if m["code"] == "SUMMARY" {
            data := m["data"].(map[string]any)
            errors = int(data["errors"].(float64))
        }
    }
    if !sawCode || errors == 0 { t.Fatalf("expected E_IO_PERMISSION and errors>0; out=%s", js.String()) }

    // Human
    var hm bytes.Buffer
    if err := runLint(&hm, dir, false, false, false); err == nil { t.Fatalf("expected error; out=%s", hm.String()) }
    s := hm.String()
    if !regexp.MustCompile(`lint: \d+ error\(s\), \d+ warning\(s\)`).MatchString(s) { t.Fatalf("missing human summary: %s", s) }
    // Ensure position formatting present
    if !regexp.MustCompile(`\(.*:\d+:\d+\)`).MatchString(s) { t.Fatalf("missing (file:line:col) formatting: %s", s) }
}

// Parity for warning promoted to error under strict: W_DUP_FUNC_DECL
func TestLint_Parity_Strict_DupFunc(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "parity_dupfunc")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    f1 := "package x\nfunc F(){}\n"
    f2 := "package x\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(srcDir, "a.ami"), []byte(f1), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.WriteFile(filepath.Join(srcDir, "b.ami"), []byte(f2), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true})
    defer setRuleToggles(RuleToggles{})

    // JSON strict
    var js bytes.Buffer
    if err := runLint(&js, dir, true, false, true); err == nil { t.Fatalf("expected error under strict; out=%s", js.String()) }
    // Human strict
    var hm bytes.Buffer
    if err := runLint(&hm, dir, false, false, true); err == nil { t.Fatalf("expected error under strict; out=%s", hm.String()) }
    if !regexp.MustCompile(`lint: \d+ error\(s\), \d+ warning\(s\)`).MatchString(hm.String()) {
        t.Fatalf("missing human summary line: %s", hm.String())
    }
}

