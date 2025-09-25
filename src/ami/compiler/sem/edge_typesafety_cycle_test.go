package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestEdgeTypeSafety_Mismatch(t *testing.T) {
    // f outputs Event<string> but edge declares type=int
    src := `package p
func f(ctx Context, ev Event<string>, st State) Event<string> { }
pipeline P { Ingress(cfg).Transform(f).Egress(in=edge.FIFO(minCapacity=0,maxCapacity=0,backpressure=block,type=int)) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var found bool
    for _, d := range res.Diagnostics { if d.Code == "E_EDGE_TYPE_MISMATCH" { found = true; break } }
    if !found { t.Fatalf("expected E_EDGE_TYPE_MISMATCH; diags=%v", res.Diagnostics) }
}

func TestEdgeTypeSafety_Match(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<string>, st State) Event<string> { }
pipeline P { Ingress(cfg).Transform(f).Egress(in=edge.FIFO(minCapacity=0,maxCapacity=0,backpressure=block,type=string)) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_EDGE_TYPE_MISMATCH" { t.Fatalf("unexpected type mismatch: %v", d) } }
}

func TestCycleDetection_ErrorWithoutPragma(t *testing.T) {
    src := `package p
pipeline A { Ingress(cfg).Egress(in=edge.Pipeline(name=B)) }
pipeline B { Ingress(cfg).Egress(in=edge.Pipeline(name=A)) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var found bool
    for _, d := range res.Diagnostics { if d.Code == "E_CYCLE_DETECTED" { found = true; break } }
    if !found { t.Fatalf("expected E_CYCLE_DETECTED; diags=%v", res.Diagnostics) }
}

func TestCycleDetection_AllowedWithPragma(t *testing.T) {
    src := `#pragma cycle allow
package p
pipeline A { Ingress(cfg).Egress(in=edge.Pipeline(name=B)) }
pipeline B { Ingress(cfg).Egress(in=edge.Pipeline(name=A)) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_CYCLE_DETECTED" { t.Fatalf("unexpected cycle diag: %v", d) } }
}
