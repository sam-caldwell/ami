package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestPipelinesDebug_MultiPath_Inputs_Listed(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\npipeline P(){ A(); B(); A -> Collect; B -> Collect; Collect edge.MultiPath(merge.Sort(\"ts\")); egress }\n"
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
    steps := pipes[0].(map[string]any)["steps"].([]any)
    // find Collect step
    var mp map[string]any
    for _, s := range steps {
        m := s.(map[string]any)
        if m["name"] == "Collect" {
            if v, ok := m["multipath"].(map[string]any); ok { mp = v; break }
        }
    }
    if mp == nil { t.Fatalf("multipath missing on Collect") }
    in, ok := mp["inputs"].([]any)
    if !ok || len(in) != 2 { t.Fatalf("inputs missing: %v", mp) }
}

