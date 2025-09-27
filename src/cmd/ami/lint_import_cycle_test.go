package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_CircularLocalImports_ErrorAndTerminate(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "cycle")
    aDir := filepath.Join(dir, "a")
    bDir := filepath.Join(dir, "b")
    if err := os.MkdirAll(aDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(bDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // minimal files to satisfy scans
    _ = os.WriteFile(filepath.Join(aDir, "main.ami"), []byte("package a\n"), 0o644)
    _ = os.WriteFile(filepath.Join(bDir, "main.ami"), []byte("package b\n"), 0o644)

    // workspace with main at ./a; a -> ./b; b -> ./a
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Linter.Options = []string{} // non-strict
    ws.Packages = workspace.PackageList{
        {Key: "main", Package: workspace.Package{Name: "a", Version: "0.0.1", Root: "./a", Import: []string{"./b"}}},
        {Key: "b", Package: workspace.Package{Name: "b", Version: "0.0.1", Root: "./b", Import: []string{"./a"}}},
    }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    err := runLint(&buf, dir, true, false, false)
    if err == nil { t.Fatalf("expected error due to cycle; out=%s", buf.String()) }
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "E_IMPORT_CYCLE" { saw = true }
    }
    if !saw { t.Fatalf("expected E_IMPORT_CYCLE; out=%s", buf.String()) }
}

