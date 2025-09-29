package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestPipelinesDebug_Collect_Instance_Disambiguation(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Two Collect instances; edges assigned to the nearest following Collect by order
    code := "package app\n" +
        "pipeline P(){\n" +
        "  A; B;\n" +
        "  A -> Collect;\n" +
        "  Collect;\n" +
        "  B -> Collect;\n" +
        "  Collect;\n" +
        "  egress\n" +
        "}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    p := filepath.Join("build", "debug", "ir", "app", "u.pipelines.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    pipes := obj["pipelines"].([]any)
    steps := pipes[0].(map[string]any)["steps"].([]any)
    // Find Collect steps and check their inputs
    var c1, c2 map[string]any
    for _, s := range steps {
        m := s.(map[string]any)
        if m["name"] == "Collect" {
            if c1 == nil { c1 = m } else { c2 = m }
        }
    }
    if c1 == nil || c2 == nil { t.Fatalf("expected two Collect steps") }
    mp1 := c1["multipath"].(map[string]any)
    mp2 := c2["multipath"].(map[string]any)
    in1 := toStrings2(mp1["inputs"]) 
    in2 := toStrings2(mp2["inputs"]) 
    if len(in1) != 1 || in1[0] != "A" { t.Fatalf("collect#1 inputs: %v", in1) }
    if len(in2) != 1 || in2[0] != "B" { t.Fatalf("collect#2 inputs: %v", in2) }
}

// helper exists in another test; re-use a distinct name here to avoid redeclare
func toStrings2(v any) []string { arr, _ := v.([]any); var out []string; for _, e := range arr { if s, ok := e.(string); ok { out = append(out, s) } }; return out }
