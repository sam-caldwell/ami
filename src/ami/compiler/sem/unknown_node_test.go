package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestAnalyzeFile_Pipeline_UnknownNode_Diagnostic(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Mystery().Egress(cfg) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics {
        if d.Code == "E_UNKNOWN_NODE" { seen = true; break }
    }
    if !seen {
        t.Fatalf("expected E_UNKNOWN_NODE diagnostic; got %+v", res.Diagnostics)
    }
}

