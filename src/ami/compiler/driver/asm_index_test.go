package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_WritesAsmIndex(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("a.ami", "package app\npipeline P(){ ingress; A -> Collect; Collect merge.Buffer(1, dropNewest); egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    p := filepath.Join("build", "debug", "asm", "app", "asm.index.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "asm.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    if obj["totalEdges"].(float64) < 1 { t.Fatalf("totalEdges missing") }
    if _, ok := obj["policyCount"].(map[string]any); !ok { t.Fatalf("policyCount missing") }
}
