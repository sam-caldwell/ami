package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_MinimalFunction_IRDebugAndNoDiags(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // simple function with var/assign/return
    code := "package app\nfunc main(a int) (int) { var x int; x = 1; return x }\n"
    fs.AddFile("unit1.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    arts, diags := Compile(ws, pkgs, Options{Debug: true})
    if len(diags) != 0 {
        b, _ := json.Marshal(diags)
        t.Fatalf("unexpected diagnostics: %s", string(b))
    }
    if len(arts.IR) != 1 { t.Fatalf("expected 1 IR file, got %d", len(arts.IR)) }
    out := arts.IR[0]
    // should be build/debug/ir/app/unit1.ir.json
    want := filepath.Join("build", "debug", "ir", "app", "unit1.ir.json")
    if out != want { t.Fatalf("unexpected path: %s", out) }
    b, err := os.ReadFile(out)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "ir.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    if obj["package"] != "app" { t.Fatalf("package: %v", obj["package"]) }

    // pipelines debug file
    pfile := filepath.Join("build", "debug", "ir", "app", "unit1.pipelines.json")
    pb, err := os.ReadFile(pfile)
    if err != nil { t.Fatalf("read pipelines: %v", err) }
    var pobj map[string]any
    if err := json.Unmarshal(pb, &pobj); err != nil { t.Fatalf("pipelines json: %v", err) }
    if pobj["schema"] != "pipelines.v1" { t.Fatalf("schema: %v", pobj["schema"]) }

    // eventmeta debug file
    emfile := filepath.Join("build", "debug", "ir", "app", "unit1.eventmeta.json")
    eb, err := os.ReadFile(emfile)
    if err != nil { t.Fatalf("read eventmeta: %v", err) }
    var emobj map[string]any
    if err := json.Unmarshal(eb, &emobj); err != nil { t.Fatalf("eventmeta json: %v", err) }
    if emobj["schema"] != "eventmeta.v1" { t.Fatalf("schema: %v", emobj["schema"]) }

    // asm debug file
    asmfile := filepath.Join("build", "debug", "asm", "app", "unit1.s")
    ab, err := os.ReadFile(asmfile)
    if err != nil { t.Fatalf("read asm: %v", err) }
    if len(ab) == 0 { t.Fatalf("asm file is empty") }
}

func TestCompile_MemSafetyDiagnostics(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // address-of operator should be flagged
    fs.AddFile("bad.ami", "package app\nfunc f(){ x = &y }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    if len(diags) == 0 { t.Fatalf("expected diagnostics for mem safety") }
}

func TestCompile_PipelineSemanticsDiagnostics(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("p.ami", "package app\npipeline P() { work; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    // Expect both start error and unknown node
    if len(diags) == 0 { t.Fatalf("expected diagnostics, got none") }
    hasStart := false
    hasUnknown := false
    for _, d := range diags { if d.Code == "E_PIPELINE_START_INGRESS" { hasStart = true }; if d.Code == "E_UNKNOWN_NODE" { hasUnknown = true } }
    if !hasStart || !hasUnknown { t.Fatalf("missing expected codes: %v", diags) }
}

func TestCompile_EdgesIndexDebug(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("g.ami", "package app\npipeline P() { A -> B; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    edges := filepath.Join("build", "debug", "asm", "app", "edges.json")
    b, err := os.ReadFile(edges)
    if err != nil { t.Fatalf("read edges: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("edges json: %v", err) }
    if obj["schema"] != "edges.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    edgesArr := obj["edges"].([]any)
    if len(edgesArr) == 0 { t.Fatalf("no edges in index") }
    e0 := edgesArr[0].(map[string]any)
    if _, ok := e0["bounded"]; !ok { t.Fatalf("bounded missing") }
    if _, ok := e0["delivery"]; !ok { t.Fatalf("delivery missing") }
}
