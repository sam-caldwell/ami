package sem

import (
    "encoding/json"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func codesJSON(diags any) string { b, _ := json.Marshal(diags); return string(b) }

func collectCodes(ds []any) []string { return nil }

func TestPipelineSemantics_KnownUnknownAndStartEnd(t *testing.T) {
    code := "package app\npipeline P() { ingress; work(); egress }\n"
    f := (&source.FileSet{}).AddFile("p.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    // Should only flag unknown node 'work', not start/end
    if len(ds) != 1 { t.Fatalf("want 1 diag, got %d: %s", len(ds), codesJSON(ds)) }
    if ds[0].Code != "E_UNKNOWN_NODE" { t.Fatalf("code: %s", ds[0].Code) }
}

func TestPipelineSemantics_MissingStart(t *testing.T) {
    code := "package app\npipeline Q() { work; egress }\n"
    f := (&source.FileSet{}).AddFile("q.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    // Expect start error and unknown node (work)
    if len(ds) != 2 { t.Fatalf("want 2 diags, got %d: %s", len(ds), codesJSON(ds)) }
    hasStart := false
    hasUnknown := false
    for _, d := range ds { if d.Code == "E_PIPELINE_START_INGRESS" { hasStart = true }; if d.Code == "E_UNKNOWN_NODE" { hasUnknown = true } }
    if !hasStart || !hasUnknown { t.Fatalf("missing expected codes: %s", codesJSON(ds)) }
}

func TestPipelineSemantics_MissingEnd(t *testing.T) {
    code := "package app\npipeline R() { ingress; work }\n"
    f := (&source.FileSet{}).AddFile("r.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    // Expect end error and unknown node (work)
    if len(ds) != 2 { t.Fatalf("want 2 diags, got %d: %s", len(ds), codesJSON(ds)) }
    hasEnd := false
    hasUnknown := false
    for _, d := range ds { if d.Code == "E_PIPELINE_END_EGRESS" { hasEnd = true }; if d.Code == "E_UNKNOWN_NODE" { hasUnknown = true } }
    if !hasEnd || !hasUnknown { t.Fatalf("missing expected codes: %s", codesJSON(ds)) }
}

func TestPipelineSemantics_EmptyPipeline(t *testing.T) {
    code := "package app\npipeline E() { }\n"
    f := (&source.FileSet{}).AddFile("e.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    // Expect both start and end errors
    if len(ds) != 2 { t.Fatalf("want 2 diags, got %d: %s", len(ds), codesJSON(ds)) }
    hasStart := false
    hasEnd := false
    for _, d := range ds { if d.Code == "E_PIPELINE_START_INGRESS" { hasStart = true }; if d.Code == "E_PIPELINE_END_EGRESS" { hasEnd = true } }
    if !hasStart || !hasEnd { t.Fatalf("missing expected codes: %s", codesJSON(ds)) }
}

func TestPipelineSemantics_IngressPositionAndDup(t *testing.T) {
    code := "package app\npipeline X() { ingress; ingress; egress }\n"
    f := (&source.FileSet{}).AddFile("x.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    // Expect duplicate ingress and position error for second ingress
    hasDup := false
    hasPos := false
    for _, d := range ds { if d.Code == "E_DUP_INGRESS" { hasDup = true }; if d.Code == "E_INGRESS_POSITION" { hasPos = true } }
    if !hasDup || !hasPos { t.Fatalf("missing dup/pos ingress: %s", codesJSON(ds)) }
}

func TestPipelineSemantics_EgressPositionAndDup(t *testing.T) {
    code := "package app\npipeline Y() { ingress; egress; egress }\n"
    f := (&source.FileSet{}).AddFile("y.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    // Expect duplicate egress and position error for first egress (not last)
    hasDup := false
    hasPos := false
    for _, d := range ds { if d.Code == "E_DUP_EGRESS" { hasDup = true }; if d.Code == "E_EGRESS_POSITION" { hasPos = true } }
    if !hasDup || !hasPos { t.Fatalf("missing dup/pos egress: %s", codesJSON(ds)) }
}

func TestPipelineSemantics_IOPermission(t *testing.T) {
    code := "package app\npipeline Z() { ingress; io.Read(\"f\"); egress }\n"
    f := (&source.FileSet{}).AddFile("z.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzePipelineSemantics(af)
    hasIO := false
    for _, d := range ds { if d.Code == "E_IO_PERMISSION" { hasIO = true } }
    if !hasIO { t.Fatalf("expected E_IO_PERMISSION: %s", codesJSON(ds)) }
}
