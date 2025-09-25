package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

// Ensure dotted worker references like pkg.Func() are accepted when the package is imported.
func TestAnalyze_WorkerImportResolution_DottedName(t *testing.T) {
    src := `package p
import "util"
func f(ctx Context, ev Event<string>, st State) Event<string> { }
pipeline P { Ingress(cfg).Transform(util.Work()).Egress(cfg) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_WORKER_UNDEFINED" {
            t.Fatalf("unexpected undefined worker diag: %+v", d)
        }
    }
}

