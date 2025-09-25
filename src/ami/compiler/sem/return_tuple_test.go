package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestReturn_Tuple_OK(t *testing.T) {
	src := `package p
func f(a int, b string) (int, string) { return a, b }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_RETURN_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %v", d)
		}
	}
}

func TestReturn_Tuple_ArityMismatch_Error(t *testing.T) {
	src := `package p
func f(a int, b string) (int, string) { return a }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_RETURN_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_RETURN_TYPE_MISMATCH; diags=%v", res.Diagnostics)
	}
}

func TestReturn_Tuple_TypeMismatch_Error(t *testing.T) {
	src := `package p
func f(a int, b string) (int, string) { return b, a }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_RETURN_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_RETURN_TYPE_MISMATCH; diags=%v", res.Diagnostics)
	}
}
