package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// When a decorated function is referenced as a worker and does not match the canonical signature,
// emit E_DECORATOR_SIGNATURE in addition to existing decorator rules.
func TestWorkers_DecoratedFunction_WithBadSignature_EmitsDecoratorSignature(t *testing.T) {
    src := "package app\n@metrics\nfunc F(a int){}\npipeline P(){ ingress; Transform(F); egress }\n"
    f := &source.File{Name: "wds.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeWorkers(af)
    var has bool
    for _, d := range ds { if d.Code == "E_DECORATOR_SIGNATURE" { has = true } }
    if !has { t.Fatalf("expected E_DECORATOR_SIGNATURE: %+v", ds) }
}

