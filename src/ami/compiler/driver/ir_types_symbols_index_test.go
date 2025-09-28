package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Build with Debug=true should write ir.types.index.json and ir.symbols.index.json with deterministic contents.
func TestCompile_IRTypesAndSymbols_Indices(t *testing.T) {
    dir := filepath.Join("build", "test", "ir_indices")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatal(err) }

    // Minimal module: one function using int and bool
    var fs source.FileSet
    fs.AddFile(filepath.Join(dir, "u.ami"), "package main\nfunc f(a int,b bool) bool { return b }\n")
    ws := DefaultWorkspace()
    pkgs := []Package{{Name: "app", Files: &fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true})
    if len(diags) != 0 { t.Fatalf("unexpected diags: %v", diags) }

    // types index
    tpath := filepath.Join("build", "debug", "ir", "app", "ir.types.index.json")
    if st, err := os.Stat(tpath); err != nil || st.IsDir() { t.Fatalf("types index missing") }
    var typesObj map[string]any
    b, _ := os.ReadFile(tpath)
    if err := json.Unmarshal(b, &typesObj); err != nil { t.Fatalf("types json: %v", err) }
    if typesObj["schema"] != "ir.types.index.v1" { t.Fatalf("schema mismatch: %v", typesObj["schema"]) }

    // symbols index
    spath := filepath.Join("build", "debug", "ir", "app", "ir.symbols.index.json")
    if st, err := os.Stat(spath); err != nil || st.IsDir() { t.Fatalf("symbols index missing") }
    var symObj map[string]any
    b, _ = os.ReadFile(spath)
    if err := json.Unmarshal(b, &symObj); err != nil { t.Fatalf("symbols json: %v", err) }
    if symObj["schema"] != "ir.symbols.index.v1" { t.Fatalf("schema mismatch: %v", symObj["schema"]) }
}

