package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestSem_Merge_Sort_NoField(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=1,backpressure=block,type=int)], merge=Sort())).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "W_MERGE_SORT_NO_FIELD" { seen = true; break } }
    if !seen { t.Fatalf("expected W_MERGE_SORT_NO_FIELD; got %+v", res.Diagnostics) }
}

func TestSem_Merge_Buffer_Tiny_Drop(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=2,maxCapacity=2,backpressure=block,type=int)], merge=Buffer(1,dropNewest))).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "W_MERGE_TINY_BUFFER" { seen = true; break } }
    if !seen { t.Fatalf("expected W_MERGE_TINY_BUFFER; got %+v", res.Diagnostics) }
}

func TestSem_Merge_UnknownAttribute(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=1,backpressure=block,type=int)], merge=Nope(1))).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "E_MERGE_ATTR_UNKNOWN" { seen = true; break } }
    if !seen { t.Fatalf("expected E_MERGE_ATTR_UNKNOWN; got %+v", res.Diagnostics) }
}

