package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestCall_GenericAmbiguous_NoArgType_Err(t *testing.T) {
	src := `package p
func g(x Event<T>) {}
func f() { g(y) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_AMBIGUOUS" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_TYPE_AMBIGUOUS; diags=%v", res.Diagnostics)
	}
}
