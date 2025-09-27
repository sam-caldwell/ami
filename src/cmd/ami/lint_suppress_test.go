package main

import (
    "bytes"
    "encoding/json"
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

