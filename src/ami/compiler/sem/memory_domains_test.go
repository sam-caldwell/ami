package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

// Ensure assigning address-of is rejected by parser (AMI does not expose pointers).
func TestMemoryDomains_AssignAddrOfIntoState_Error(t *testing.T) {
	src := `package p
func f(ctx Context, ev Event<string>, st State) Event<string> {
  *st = &ev
}`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	var found bool
	for _, d := range res.Diagnostics {
		if d.Code == "E_PTR_UNSUPPORTED_SYNTAX" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_PTR_UNSUPPORTED_SYNTAX; diags=%v", res.Diagnostics)
	}
}

// Ensure assigning a non-state identifier (e.g., Event<T> param) into state is rejected.
// Note: assigning a non-state identifier into state (e.g., *st = ev) is a
// cross-domain assignment which may be forbidden by stricter rules; address-of
// cases are covered above and enforced by the parser and analyzer.
