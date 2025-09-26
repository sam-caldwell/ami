package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

// Ensure legacy 3-param worker signature is accepted and does not raise E_WORKER_SIGNATURE.
func TestWorkers_LegacySignature_Accepted_WithDeprecatedInfo(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<int>, st State) Event<int> { }
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_WORKER_SIGNATURE" {
            t.Fatalf("unexpected signature error for legacy worker: %+v", d)
        }
    }
}

