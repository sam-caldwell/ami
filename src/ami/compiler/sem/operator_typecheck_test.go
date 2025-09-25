package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestOperators_Add_IntAndString_Mismatch(t *testing.T) {
	src := `package p
func f(a int, b string) { a + b }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_TYPE_MISMATCH; diags=%v", res.Diagnostics)
	}
}

func TestOperators_Add_IntAndInt_OK(t *testing.T) {
	src := `package p
func f(a int, b int) { a + b }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %v", d)
		}
	}
}

func TestOperators_Compare_Eq_Mismatch(t *testing.T) {
	src := `package p
func f(a int, b string) { a == b }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_TYPE_MISMATCH; diags=%v", res.Diagnostics)
	}
}

func TestOperators_Compare_Order_Int_OK(t *testing.T) {
	src := `package p
func f(a int, b int) { a < b }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %v", d)
		}
	}
}
