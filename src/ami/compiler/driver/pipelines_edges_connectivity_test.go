package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestPipelinesDebug_Edges_And_Connectivity(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\npipeline P(){ ingress; A(); B(); egress; A -> B; }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    p := filepath.Join("build", "debug", "ir", "app", "u.pipelines.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    pipes := obj["pipelines"].([]any)
    if len(pipes) == 0 { t.Fatalf("no pipelines") }
    ent := pipes[0].(map[string]any)
    // edges present
    ed, ok := ent["edges"].([]any)
    if !ok || len(ed) == 0 { t.Fatalf("edges missing: %v", ent["edges"]) }
    // connectivity present
    conn, ok := ent["connectivity"].(map[string]any)
    if !ok { t.Fatalf("connectivity missing") }
    if _, ok := conn["hasEdges"]; !ok { t.Fatalf("hasEdges missing") }
    if _, ok := conn["ingressToEgress"]; !ok { t.Fatalf("ingressToEgress missing") }
}

