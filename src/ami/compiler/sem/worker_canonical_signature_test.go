package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "testing"
)

// Ensure canonical worker signature is accepted without E_WORKER_SIGNATURE
// and without legacy deprecation notice.
func TestWorkers_CanonicalSignature_Accepted_NoDeprecated(t *testing.T) {
    src := `package p
func f(ev Event<int>) (Event<int>, error) { }
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_WORKER_SIGNATURE" {
            t.Fatalf("unexpected signature error: %+v", d)
        }
        if d.Code == "W_WORKER_SIGNATURE_DEPRECATED" {
            t.Fatalf("canonical signature should not be marked deprecated: %+v", d)
        }
    }
}

