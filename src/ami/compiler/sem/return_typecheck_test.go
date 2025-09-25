package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestReturn_IntAdd_OK(t *testing.T) {
	src := `package p
func f(a int, b int) int { return a + b }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_RETURN_TYPE_MISMATCH" || d.Code == "E_TYPE_MISMATCH" || d.Code == "E_TYPE_UNINFERRED" {
			t.Fatalf("unexpected diag: %v", d)
		}
	}
}

func TestReturn_Mismatch_Error(t *testing.T) {
	src := `package p
func f(a int) string { return a }`
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

func TestReturn_GenericUnify_OK(t *testing.T) {
	src := `package p
func f(x Event<string>) Event<T> { return x }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_RETURN_TYPE_MISMATCH" || d.Code == "E_TYPE_UNINFERRED" {
			t.Fatalf("unexpected diag: %v", d)
		}
	}
}

func TestReturn_Generic_Uninferred_Error(t *testing.T) {
	src := `package p
func f(x Event<T>) Event<T> { return x }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_UNINFERRED" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_TYPE_UNINFERRED; diags=%v", res.Diagnostics)
	}
}
