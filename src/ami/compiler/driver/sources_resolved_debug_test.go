package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_ResolvedSourcesDebug_Written(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("unit1.ami", "package app\nimport alpha >= v1.0.0\nfunc F(){}\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    p := filepath.Join("build", "debug", "source", "resolved.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v", e) }
    if obj["schema"] != "sources.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    if _, ok := obj["units"].([]any); !ok { t.Fatalf("units missing or wrong type: %T", obj["units"]) }
}

