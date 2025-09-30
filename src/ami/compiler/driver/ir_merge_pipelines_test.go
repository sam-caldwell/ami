package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestIR_Pipelines_MergePlan_Encoded(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n" +
        "pipeline P(){ A; Collect merge.Sort(\"k\", desc), merge.Buffer(10, dropOldest), merge.Stable(); A -> Collect; egress }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    p, ok := obj["pipelines"].([]any)
    if !ok || len(p) == 0 { t.Fatalf("pipelines missing: %v", obj) }
    first := p[0].(map[string]any)
    cols, ok := first["collect"].([]any)
    if !ok || len(cols) == 0 { t.Fatalf("collect missing: %v", first) }
    c0 := cols[0].(map[string]any)
    merge, ok := c0["merge"].(map[string]any)
    if !ok { t.Fatalf("merge missing: %v", c0) }
    if merge["stable"] != true { t.Fatalf("stable: %v", merge["stable"]) }
    sortArr, _ := merge["sort"].([]any)
    if len(sortArr) == 0 { t.Fatalf("sort not encoded: %v", merge) }
    s0 := sortArr[0].(map[string]any)
    if s0["field"] != "k" || s0["order"] != "desc" { t.Fatalf("sort wrong: %v", s0) }
    buf, _ := merge["buffer"].(map[string]any)
    if buf["capacity"].(float64) != 10 || buf["policy"].(string) != "dropoldest" && buf["policy"].(string) != "dropOldest" {
        t.Fatalf("buffer wrong: %v", buf)
    }
}

