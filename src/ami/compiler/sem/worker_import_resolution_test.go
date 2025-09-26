package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

// Ensure dotted worker references like pkg.Func() are accepted when the package is imported.
func TestAnalyze_WorkerImportResolution_DottedName(t *testing.T) {
	src := `package p
import "util"
func f(ev Event<string>) (Event<string>, error) { }
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
