package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestWriteIRRAIIDebug_WritesPerUnitTrace(t *testing.T) {
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ var x string; release(x); defer release(x) }\n"
    fs.AddFile("r.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(workspace.Workspace{}, pkgs, Options{Debug: true})
    // Expect RAII debug file alongside IR
    p := filepath.Join("build", "debug", "ir", "app")
    // unit name derived from file base sans extension
    f := filepath.Join(p, "r.raii.json")
    b, err := os.ReadFile(f)
    if err != nil { t.Fatalf("read raii: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "ir.raii.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    // ensure functions array present and contains our function
    arr, ok := obj["functions"].([]any)
    if !ok || len(arr) == 0 { t.Fatalf("missing functions in raii json") }
}

