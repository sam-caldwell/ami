package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_LocalImport_UndeclaredPackage_Warns(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "undeclared")
    sub := filepath.Join(dir, "lib")
    if err := os.MkdirAll(sub, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create workspace with main importing ./lib but without declaring a package for ./lib
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Import = []string{"./lib"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // JSON mode returns nil unless errors are present; warnings are OK
    }
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_IMPORT_LOCAL_UNDECLARED" { saw = true }
    }
    if !saw { t.Fatalf("expected W_IMPORT_LOCAL_UNDECLARED; out=%s", buf.String()) }
}

