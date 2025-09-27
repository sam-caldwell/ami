package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify --rules can mix glob and regex and filter correctly.
func TestLint_Rules_MixedGlobAndRegex(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "rules_mixed")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\n// TODO: note\nfunc F(){ var bad_name }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    // Allow only W_TODO (regex) and W_IDENT_* (glob). Expect to exclude others.
    setLintOptions(LintOptions{Rules: []string{"/^(W_TODO)$/", "W_IDENT_*"}})
    defer setLintOptions(LintOptions{})
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { /* warnings allowed */ }
    dec := json.NewDecoder(&buf)
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "SUMMARY" { break }
        code := m["code"].(string)
        if !(code == "W_TODO" || code == "W_IDENT_UNDERSCORE") {
            t.Fatalf("unexpected code after filter: %s\nout=%s", code, buf.String())
        }
    }
}

