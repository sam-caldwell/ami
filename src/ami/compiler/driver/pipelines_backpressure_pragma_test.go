package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestPipelinesDebug_BackpressurePragma_DefaultDelivery(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n#pragma backpressure policy=dropOldest\npipeline P(){ Alpha() }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.pipelines.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read pipelines: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    pipes := obj["pipelines"].([]any)
    p0 := pipes[0].(map[string]any)
    steps := p0["steps"].([]any)
    s0 := steps[0].(map[string]any)
    edge := s0["edge"].(map[string]any)
    if edge["delivery"] != "bestEffort" { t.Fatalf("expected bestEffort, got %v", edge["delivery"]) }
}

