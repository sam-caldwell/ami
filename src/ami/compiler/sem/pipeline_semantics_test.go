package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestAnalyzeFile_Pipeline_HappyPath_StartIngress_EndEgress(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    if len(res.Diagnostics) != 0 {
        t.Fatalf("expected no diagnostics; got %v", res.Diagnostics)
    }
}

func TestAnalyzeFile_Pipeline_MustStartWithIngress(t *testing.T) {
    src := `package p
pipeline P { Transform(f).Egress(cfg) }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    if len(res.Diagnostics) == 0 {
        t.Fatalf("expected diagnostics for missing ingress")
    }
    found := false
    for _, d := range res.Diagnostics {
        if d.Code == "E_PIPELINE_START_INGRESS" { found = true; break }
    }
    if !found { t.Fatalf("expected E_PIPELINE_START_INGRESS; diags=%v", res.Diagnostics) }
}

func TestAnalyzeFile_Pipeline_MustEndWithEgress(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Transform(f) }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    if len(res.Diagnostics) == 0 {
        t.Fatalf("expected diagnostics for missing egress")
    }
    found := false
    for _, d := range res.Diagnostics {
        if d.Code == "E_PIPELINE_END_EGRESS" { found = true; break }
    }
    if !found { t.Fatalf("expected E_PIPELINE_END_EGRESS; diags=%v", res.Diagnostics) }
}

