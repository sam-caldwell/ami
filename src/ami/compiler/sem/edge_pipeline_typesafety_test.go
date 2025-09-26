package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "testing"
)

func TestEdgePipelineTypeSafety_Match(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<int>, st State) Event<int> { }
pipeline X { Ingress(cfg).Transform(f).Egress() }
pipeline Y { Ingress(cfg).Egress(in=edge.Pipeline(name=X,type=int)) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_EDGE_PIPE_TYPE_MISMATCH" || d.Code == "E_EDGE_PIPE_NOT_FOUND" {
            t.Fatalf("unexpected diag: %v", d)
        }
    }
}

func TestEdgePipelineTypeSafety_Mismatch(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<int>, st State) Event<int> { }
pipeline X { Ingress(cfg).Transform(f).Egress() }
pipeline Y { Ingress(cfg).Egress(in=edge.Pipeline(name=X,type=string)) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var found bool
    for _, d := range res.Diagnostics {
        if d.Code == "E_EDGE_PIPE_TYPE_MISMATCH" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected E_EDGE_PIPE_TYPE_MISMATCH; diags=%v", res.Diagnostics)
    }
}

func TestEdgePipelineTypeSafety_UnknownPipeline(t *testing.T) {
    src := `package p
pipeline Y { Ingress(cfg).Egress(in=edge.Pipeline(name=Z,type=int)) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var found bool
    for _, d := range res.Diagnostics {
        if d.Code == "E_EDGE_PIPE_NOT_FOUND" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected E_EDGE_PIPE_NOT_FOUND; diags=%v", res.Diagnostics)
    }
}

