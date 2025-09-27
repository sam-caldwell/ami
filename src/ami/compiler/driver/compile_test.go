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
    // Assert positions and file are populated
    hasPos := false
    for _, d := range diags {
        if d.Pos != nil && d.File != "" && d.Pos.Line > 0 && d.Pos.Column > 0 {
            hasPos = true
            break
        }
    }
    if !hasPos { t.Fatalf("expected at least one diagnostic with file and position; got: %+v", diags) }
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
    // Positions present
    for _, d := range diags {
        if d.Code == "E_PIPELINE_START_INGRESS" || d.Code == "E_UNKNOWN_NODE" {
            if d.File == "" || d.Pos == nil || d.Pos.Line <= 0 || d.Pos.Column <= 0 {
                t.Fatalf("expected file/pos on diag: %+v", d)
            }
        }
    }
}

func TestCompile_PipelinesDebug_AttrsRaw(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("pp.ami", "package app\npipeline P(){ Alpha() edge.MultiPath(merge.Stable()) ; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    pfile := filepath.Join("build", "debug", "ir", "app", "pp.pipelines.json")
    b, err := os.ReadFile(pfile)
    if err != nil { t.Fatalf("read pipelines: %v", err) }
    var pobj map[string]any
    if err := json.Unmarshal(b, &pobj); err != nil { t.Fatalf("json: %v", err) }
    pipes, _ := pobj["pipelines"].([]any)
    if len(pipes) != 1 { t.Fatalf("pipes len: %d", len(pipes)) }
    p0 := pipes[0].(map[string]any)
    steps, _ := p0["steps"].([]any)
    if len(steps) == 0 { t.Fatalf("steps empty") }
    s0 := steps[0].(map[string]any)
    attrs, ok := s0["attrs"].([]any)
    if !ok || len(attrs) != 1 { t.Fatalf("attrs: %#v", s0["attrs"]) }
    a0 := attrs[0].(map[string]any)
    if a0["name"] != "edge.MultiPath" { t.Fatalf("attr name: %v", a0["name"]) }
    av, _ := a0["args"].([]any)
    if len(av) != 1 || av[0].(string) != "merge.Stable()" { t.Fatalf("attr args: %#v", av) }
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

func TestCompile_IRTyping_CallResultType(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc G() (int) { return 1 }\nfunc F() (int) { return G() }\n"
    fs.AddFile("call.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    arts, _ := Compile(ws, pkgs, Options{Debug: true})
    // locate F unit IR and inspect return type
    if len(arts.IR) == 0 { t.Fatalf("no IR emitted") }
    // read any IR; in this scaffold it's per-unit; pick the first
    b, err := os.ReadFile(arts.IR[0])
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    // find F
    var f map[string]any
    for _, it := range fns {
        m := it.(map[string]any)
        if m["name"] == "F" { f = m; break }
    }
    if f == nil { t.Fatalf("function F not found") }
    blks := f["blocks"].([]any)
    blk := blks[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    last := instrs[len(instrs)-1].(map[string]any)
    if last["op"] != "RETURN" { t.Fatalf("last op: %v", last["op"]) }
    vals := last["values"].([]any)
    v0 := vals[0].(map[string]any)
    if v0["type"] != "int" { t.Fatalf("return type: %v", v0["type"]) }
}

func TestCompile_EdgesIndexDebug(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("g.ami", "package app\npipeline P() { ingress; A type(\"X\"); A -> Collect; Collect type(\"X\"), merge.Buffer(10, dropNewest), merge.Sort(\"ts\"); MultiPath(u,v); egress }\n")
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
    if e0["bounded"] != true || e0["delivery"] != "bestEffort" { t.Fatalf("edge derivation wrong: %+v", e0) }
    if e0["type"] != "X" { t.Fatalf("edge type not propagated: %+v", e0) }
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
    code := "package app\nimport alpha >= v1.0.0\nfunc F<T any>(a T) (R) { return a }\npipeline P(){ Alpha(\"x\", y) edge.MultiPath(merge.Sort(\"ts\"), merge.Stable()) }\n"
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
    // imports content
    imps, ok := obj["imports"].([]any)
    if !ok || len(imps) != 1 { t.Fatalf("imports: %T len=%d", obj["imports"], len(imps)) }
    imp0, _ := imps[0].(map[string]any)
    if imp0["path"] != "alpha" || imp0["constraint"] != ">= v1.0.0" { t.Fatalf("import0: %+v", imp0) }
    // funcs content
    fns, ok := obj["funcs"].([]any)
    if !ok || len(fns) != 1 { t.Fatalf("funcs: %T len=%d", obj["funcs"], len(fns)) }
    fn0, _ := fns[0].(map[string]any)
    if fn0["name"] != "F" { t.Fatalf("func name: %v", fn0["name"]) }
    tps, _ := fn0["typeParams"].([]any)
    if len(tps) != 1 { t.Fatalf("typeParams len: %d", len(tps)) }
    tp0, _ := tps[0].(map[string]any)
    if tp0["name"] != "T" || tp0["constraint"] != "any" { t.Fatalf("typeParam0: %+v", tp0) }
    // pipelines content
    pipes, ok := obj["pipelines"].([]any)
    if !ok || len(pipes) != 1 { t.Fatalf("pipelines: %T len=%d", obj["pipelines"], len(pipes)) }
    p0, _ := pipes[0].(map[string]any)
    if p0["name"] != "P" { t.Fatalf("pipe name: %v", p0["name"]) }
    steps, _ := p0["steps"].([]any)
    if len(steps) != 1 { t.Fatalf("steps len: %d", len(steps)) }
    s0, _ := steps[0].(map[string]any)
    if s0["name"] != "Alpha" { t.Fatalf("step name: %v", s0["name"]) }
    if args, ok := s0["args"].([]any); !ok || len(args) != 2 || args[0].(string) != "x" || args[1].(string) != "y" {
        t.Fatalf("step args: %#v", s0["args"])
    }
    if attrs, ok := s0["attrs"].([]any); !ok || len(attrs) != 1 {
        t.Fatalf("attrs: %#v", s0["attrs"])
    } else {
        a0 := attrs[0].(map[string]any)
        if a0["name"] != "edge.MultiPath" { t.Fatalf("attr name: %v", a0["name"]) }
        av, _ := a0["args"].([]any)
        if len(av) != 2 || av[0].(string) != "merge.Sort(…)" || av[1].(string) != "merge.Stable()" {
            t.Fatalf("attr args: %#v", av)
        }
    }
}
