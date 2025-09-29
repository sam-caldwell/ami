package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestPipelinesDebug_Concurrency_Header(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n#pragma concurrency:workers 4\n#pragma concurrency:schedule fair\npipeline P(){ ingress; egress }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    p := filepath.Join("build", "debug", "ir", "app", "u.pipelines.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    conc, ok := obj["concurrency"].(map[string]any)
    if !ok { t.Fatalf("missing concurrency header: %v", obj) }
    if conc["workers"].(float64) != 4 || conc["schedule"].(string) != "fair" { t.Fatalf("bad concurrency: %+v", conc) }
}

