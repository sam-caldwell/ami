package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestIR_Function_IncludesDecorators(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n@trace\n@deprecated(\"x\")\nfunc F(){}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns, ok := obj["functions"].([]any)
    if !ok || len(fns) != 1 { t.Fatalf("functions: %T len=%d", obj["functions"], len(fns)) }
    f0 := fns[0].(map[string]any)
    decs, ok := f0["decorators"].([]any)
    if !ok || len(decs) != 2 { t.Fatalf("decorators: %#v", f0["decorators"]) }
}

