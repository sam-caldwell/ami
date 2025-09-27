package main

import (
    "bytes"
    "encoding/json"
    "strings"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_ConfigSuppression_ByPathPrefix(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "suppress")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Trigger two rules under ./src
    content := "bad_ident := 1\nUNKNOWN_IDENT\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if ws.Toolchain.Linter.Suppress == nil { ws.Toolchain.Linter.Suppress = map[string][]string{} }
    ws.Toolchain.Linter.Suppress["./src"] = []string{"W_IDENT_UNDERSCORE"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // ignore
    }
    dec := json.NewDecoder(&buf)
    var sawUnknown, sawIdent bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_UNKNOWN_IDENT" { sawUnknown = true }
        if m["code"] == "W_IDENT_UNDERSCORE" { sawIdent = true }
    }
    if !sawUnknown { t.Fatalf("expected W_UNKNOWN_IDENT present") }
    if sawIdent { t.Fatalf("expected W_IDENT_UNDERSCORE suppressed by config") }
}

func TestLint_ConfigSuppression_NestedPrefixes(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "suppress_nested")
    root := filepath.Join(dir, "src")
    sub := filepath.Join(root, "sub")
    if err := os.MkdirAll(sub, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(root, "a.ami"), []byte("UNKNOWN_IDENT\n"), 0o644); err != nil { t.Fatalf("write a: %v", err) }
    if err := os.WriteFile(filepath.Join(sub, "b.ami"), []byte("UNKNOWN_IDENT\n"), 0o644); err != nil { t.Fatalf("write b: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if ws.Toolchain.Linter.Suppress == nil { ws.Toolchain.Linter.Suppress = map[string][]string{} }
    // Suppress only under ./src/sub
    ws.Toolchain.Linter.Suppress["./src/sub"] = []string{"W_UNKNOWN_IDENT"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { /* allow */ }
    dec := json.NewDecoder(&buf)
    var sawA, sawB bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "W_UNKNOWN_IDENT" {
            if f, ok := m["file"].(string); ok {
                if strings.HasSuffix(f, filepath.Join("src", "a.ami")) { sawA = true }
                if strings.HasSuffix(f, filepath.Join("src", "sub", "b.ami")) { sawB = true }
            }
        }
    }
    if !sawA { t.Fatalf("expected unsuppressed warning in parent dir (a.ami)") }
    if sawB { t.Fatalf("expected suppressed warning in nested dir (b.ami)") }
}
