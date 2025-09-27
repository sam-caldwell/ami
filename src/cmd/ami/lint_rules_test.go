package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_RuleMapping_OffAndInfo(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "rules")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Name = "bad_name"              // W_PKG_NAME_STYLE
    ws.Packages[0].Package.Import = []string{"bad space"} // W_IMPORT_SYNTAX
    if ws.Toolchain.Linter.Rules == nil { ws.Toolchain.Linter.Rules = map[string]string{} }
    ws.Toolchain.Linter.Rules["W_IMPORT_SYNTAX"] = "off"
    ws.Toolchain.Linter.Rules["W_PKG_NAME_STYLE"] = "info"
    ws.Toolchain.Linter.Options = []string{} // non-strict
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { t.Fatalf("runLint: %v", err) }
    // decode NDJSON and verify codes/levels
    dec := json.NewDecoder(&buf)
    var hasPkgInfo, hasImportSyntax bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json decode: %v", err) }
        if m["code"] == "SUMMARY" { break }
        if m["code"] == "W_PKG_NAME_STYLE" && m["level"] == "info" { hasPkgInfo = true }
        if m["code"] == "W_IMPORT_SYNTAX" { hasImportSyntax = true }
    }
    if !hasPkgInfo { t.Fatalf("expected W_PKG_NAME_STYLE as info") }
    if hasImportSyntax { t.Fatalf("expected W_IMPORT_SYNTAX suppressed by rules") }
}

func TestLint_PragmaDisable_UnknownIdent(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "pragma")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create a file with pragma to disable unknown ident, and one occurrence
    content := "#pragma lint:disable W_UNKNOWN_IDENT\nUNKNOWN_IDENT\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{} // non-strict
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { t.Fatalf("runLint: %v", err) }
    // Expect no W_UNKNOWN_IDENT records
    dec := json.NewDecoder(&buf)
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_UNKNOWN_IDENT" { t.Fatalf("pragma failed to disable W_UNKNOWN_IDENT: %s", buf.String()) }
    }
}

