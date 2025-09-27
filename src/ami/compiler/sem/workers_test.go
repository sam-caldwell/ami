package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestWorkers_Undefined(t *testing.T) {
    code := "package app\n" +
        "pipeline P(){ ingress; Transform(F); egress }\n"
    f := (&source.FileSet{}).AddFile("w1.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    found := false
    for _, d := range ds { if d.Code == "E_WORKER_UNDEFINED" { found = true } }
    if !found { t.Fatalf("expected E_WORKER_UNDEFINED: %+v", ds) }
}

func TestWorkers_BadSignature(t *testing.T) {
    code := "package app\n" +
        "func F(a int){}\n" +
        "pipeline P(){ ingress; Transform(F); egress }\n"
    f := (&source.FileSet{}).AddFile("w2.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    found := false
    for _, d := range ds { if d.Code == "E_WORKER_SIGNATURE" { found = true } }
    if !found { t.Fatalf("expected E_WORKER_SIGNATURE: %+v", ds) }
}

func TestWorkers_GoodSignature(t *testing.T) {
    code := "package app\n" +
        "func F(ev Event<T>) (Event<U>, error) { return ev, nil }\n" +
        "pipeline P(){ ingress; Transform(F); egress }\n"
    f := (&source.FileSet{}).AddFile("w3.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    for _, d := range ds { if d.Code == "E_WORKER_SIGNATURE" || d.Code == "E_WORKER_UNDEFINED" { t.Fatalf("unexpected: %+v", ds) } }
}

func TestWorkers_Signature_Checked_WithDecorators(t *testing.T) {
    code := "package app\n" +
        "@metrics\nfunc F(a int){}\n" +
        "pipeline P(){ ingress; Transform(F); egress }\n"
    f := (&source.FileSet{}).AddFile("w5.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    found := false
    for _, d := range ds { if d.Code == "E_WORKER_SIGNATURE" { found = true } }
    if !found { t.Fatalf("expected E_WORKER_SIGNATURE with decorator present: %+v", ds) }
}

func TestWorkers_DottedImport_SuppressesUndefined(t *testing.T) {
    code := "package app\n" +
        "import alpha\n" +
        "pipeline P(){ ingress; Transform(alpha.F); egress }\n"
    f := (&source.FileSet{}).AddFile("w4.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    for _, d := range ds {
        if d.Code == "E_WORKER_UNDEFINED" || d.Code == "E_WORKER_SIGNATURE" {
            t.Fatalf("unexpected worker diag for dotted import: %+v", ds)
        }
    }
}
