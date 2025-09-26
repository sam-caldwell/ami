package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "testing"
)

func TestSem_MultiPath_OnlyOnCollect(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Transform(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=1,backpressure=block,type=int)])).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "E_MP_ONLY_COLLECT" { seen = true; break } }
    if !seen { t.Fatalf("expected E_MP_ONLY_COLLECT; got %+v", res.Diagnostics) }
}

func TestSem_MultiPath_FirstInput_FIFO_Required(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.Pipeline(name=X,minCapacity=0,maxCapacity=0,backpressure=dropNewest,type=int)])).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "E_MP_INPUT0_KIND" { seen = true; break } }
    if !seen { t.Fatalf("expected E_MP_INPUT0_KIND; got %+v", res.Diagnostics) }
}

func TestSem_MultiPath_TypeMismatch(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=1,backpressure=block,type=int), edge.Pipeline(name=X,minCapacity=0,maxCapacity=0,backpressure=dropNewest,type=string)])).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "E_MP_INPUT_TYPE_MISMATCH" { seen = true; break } }
    if !seen { t.Fatalf("expected E_MP_INPUT_TYPE_MISMATCH; got %+v", res.Diagnostics) }
}

func TestSem_MultiPath_Valid_Minimal(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=1,backpressure=block,type=int), edge.Pipeline(name=X,minCapacity=0,maxCapacity=0,backpressure=dropNewest,type=int)], merge=Sort("event.ts","asc"))).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_MP_ONLY_COLLECT" || d.Code == "E_MP_INPUT0_KIND" || d.Code == "E_MP_INPUT_TYPE_MISMATCH" || d.Code == "E_MP_INVALID" || d.Code == "E_MP_INPUTS_EMPTY" {
            t.Fatalf("unexpected multipath error: %+v", d)
        }
    }
}

func TestSem_MultiPath_InvalidMergeName(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=1,backpressure=block,type=int)], merge=Nope())).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics {
        if d.Code == "E_MP_MERGE_INVALID" || d.Code == "E_MERGE_ATTR_UNKNOWN" {
            seen = true
            break
        }
    }
    if !seen { t.Fatalf("expected E_MP_MERGE_INVALID or E_MERGE_ATTR_UNKNOWN; got %+v", res.Diagnostics) }
}
