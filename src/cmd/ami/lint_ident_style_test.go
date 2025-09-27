package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_IdentifierUnderscore_Warns(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "ident_style")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Contains an identifier with underscore
    content := "bad_ident := 1\n_ := 2\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // warnings only; no error expected in JSON mode
    }
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_IDENT_UNDERSCORE" { saw = true }
    }
    if !saw { t.Fatalf("expected W_IDENT_UNDERSCORE; out=%s", buf.String()) }
}

func TestLint_IdentifierUnderscore_PragmaDisable(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "ident_style_pragma")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "#pragma lint:disable W_IDENT_UNDERSCORE\nbad_ident := 1\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // ignore
    }
    dec := json.NewDecoder(&buf)
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_IDENT_UNDERSCORE" { t.Fatalf("pragma failed to suppress identifier warning: %s", buf.String()) }
    }
}

