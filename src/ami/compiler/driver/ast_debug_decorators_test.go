package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestWriteASTDebug_IncludesDecorators(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n@trace\n@deprecated(\"msg\")\nfunc F() {}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ast", "app", "u.ast.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ast: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    funcs, ok := obj["funcs"].([]any)
    if !ok || len(funcs) != 1 { t.Fatalf("funcs: %T len=%d", obj["funcs"], len(funcs)) }
    f0 := funcs[0].(map[string]any)
    decos, ok := f0["decorators"].([]any)
    if !ok || len(decos) != 2 { t.Fatalf("decorators: %#v", f0["decorators"]) }
    d0 := decos[0].(map[string]any)
    d1 := decos[1].(map[string]any)
    if d0["name"] != "trace" || d1["name"] != "deprecated" { t.Fatalf("order: %#v", decos) }
    d0 := decos[0].(map[string]any)
    d1 := decos[1].(map[string]any)
    if d0["name"] != "trace" { t.Fatalf("d0 name: %v", d0["name"]) }
    if d1["name"] != "deprecated" { t.Fatalf("d1 name: %v", d1["name"]) }
    if args, _ := d1["args"].([]any); len(args) != 1 || args[0].(string) == "" { t.Fatalf("d1 args: %#v", d1["args"]) }
}
