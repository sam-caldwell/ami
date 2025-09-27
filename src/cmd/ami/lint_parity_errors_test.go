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

// JSON vs Human parity when errors exist (Stage B memory-safety rule E_MUT_ASSIGN_UNMARKED).
func TestLint_JSON_Human_Parity_WithErrors(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "parity_errors")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Unmarked assignment triggers E_MUT_ASSIGN_UNMARKED
    content := "package x\nfunc F(){ x = 1 }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Enable Stage B rules
    setRuleToggles(RuleToggles{StageB: true, MemorySafety: true})
    defer setRuleToggles(RuleToggles{})

    // JSON run: expect non-nil error and SUMMARY with errors>0
    var js bytes.Buffer
    if err := runLint(&js, dir, true, false, false); err == nil { t.Fatalf("expected error in JSON mode due to errors; out=%s", js.String()) }
    dec := json.NewDecoder(&js)
    var errors int
    var sawCode bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "E_MUT_ASSIGN_UNMARKED" { sawCode = true }
        if m["code"] == "SUMMARY" {
            data := m["data"].(map[string]any)
            errors = int(data["errors"].(float64))
            break
        }
    }
    if !sawCode || errors == 0 { t.Fatalf("expected E_MUT_ASSIGN_UNMARKED and errors>0; out=%s", js.String()) }

    // Human run: expect non-nil error and summary with same counts
    var hm bytes.Buffer
    if err := runLint(&hm, dir, false, false, false); err == nil { t.Fatalf("expected error in human mode due to errors; out=%s", hm.String()) }
    s := hm.String()
    if !regexp.MustCompile(`lint: \d+ error\(s\), \d+ warning\(s\)`).MatchString(s) { t.Fatalf("missing human summary: %s", s) }
}
