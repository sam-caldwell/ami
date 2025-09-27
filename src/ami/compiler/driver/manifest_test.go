package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_WritesBuildManifest(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("m.ami", "package app\nfunc F(){}\npipeline P(){ ingress; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    p := filepath.Join("build", "debug", "manifest.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "manifest.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    pkgsArr, ok := obj["packages"].([]any)
    if !ok || len(pkgsArr) == 0 { t.Fatalf("packages array missing or empty") }
}

