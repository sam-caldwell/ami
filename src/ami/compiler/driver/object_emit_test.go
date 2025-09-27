package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_ObjectEmission_Index(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("o.ami", "package app\nfunc F(){}\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    // object file exists
    of := filepath.Join("build", "obj", "app", "o.o")
    if _, err := os.Stat(of); err != nil { t.Fatalf("object not found: %v", err) }
    // index exists and lists unit
    idxp := filepath.Join("build", "obj", "app", "index.json")
    b, err := os.ReadFile(idxp)
    if err != nil { t.Fatalf("read index: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v", e) }
    if obj["schema"] != "objindex.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    units, ok := obj["units"].([]any)
    if !ok || len(units) == 0 { t.Fatalf("units invalid: %T len=%d", obj["units"], len(units)) }
}

