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

func TestCompile_IRTyping_ReduceAnyOnReturn(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("ty.ami", "package app\nfunc F(){ var x int; return x }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    arts, _ := Compile(ws, pkgs, Options{Debug: true})
    if len(arts.IR) == 0 { t.Fatalf("no IR emitted") }
    b, err := os.ReadFile(arts.IR[0])
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    fn := fns[0].(map[string]any)
    blks := fn["blocks"].([]any)
    blk := blks[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    // locate last instruction (RETURN)
    last := instrs[len(instrs)-1].(map[string]any)
    if last["op"] != "RETURN" { t.Fatalf("last op not RETURN: %v", last["op"]) }
    vals := last["values"].([]any)
    if len(vals) != 1 { t.Fatalf("want 1 return value") }
    v0 := vals[0].(map[string]any)
    if v0["type"] != "int" { t.Fatalf("want return type int, got %v", v0["type"]) }
}

func TestCompile_EdgesIndexDebug(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("g.ami", "package app\npipeline P() { A -> Collect; Collect merge.Buffer(10, dropNewest), merge.Sort(\"ts\"); MultiPath(u,v); egress }\n")
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
    // multipath collect snapshot present
    if obj["collect"] == nil { t.Fatalf("collect snapshot missing") }
    col := obj["collect"].([]any)
    if len(col) == 0 { t.Fatalf("collect entries empty") }
}

func TestCompile_SourcesDebug_ImportsDetailed(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nimport alpha >= v1.2.3\nimport \"beta\"\nfunc main(){}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "source", "app", "u.sources.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read sources: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "sources.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    imp, ok := obj["importsDetailed"].([]any)
    if !ok || len(imp) != 2 { t.Fatalf("importsDetailed: %T len=%d", obj["importsDetailed"], len(imp)) }
}

func TestCompile_ASTDebug_FileExistsAndSchema(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nimport alpha >= v1.0.0\nfunc F<T any>(a T) (R) { return a }\npipeline P(){ Alpha() }\n"
    fs.AddFile("u2.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ast", "app", "u2.ast.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ast: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "ast.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    if obj["package"] != "app" { t.Fatalf("package: %v", obj["package"]) }
}
