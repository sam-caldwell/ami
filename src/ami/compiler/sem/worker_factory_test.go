package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestFactory_InvalidSignature_Diagnostic(t *testing.T) {
    src := `package p
func NewWorker(a int) int {}
pipeline P { Ingress(cfg).Transform(NewWorker(cfg)).Egress(cfg) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "E_WORKER_SIGNATURE" { seen = true; break } }
    if !seen { t.Fatalf("expected E_WORKER_SIGNATURE for invalid factory; got %+v", res.Diagnostics) }
}

// Note: Bare factory references (without call syntax) are not currently
// standardized; tests focus on explicit factory calls.
