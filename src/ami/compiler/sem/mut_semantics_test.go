package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestMutability_UnmarkedAssignment_Error(t *testing.T) {
	src := `package p
func f() { x = 1 }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_MUT_ASSIGN_UNMARKED" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_MUT_ASSIGN_UNMARKED; diags=%v", res.Diagnostics)
	}
}

func TestMutability_StarLHS_AllowsAssignment_OK(t *testing.T) {
	src := `package p
func f() { *x = 1 }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_MUT_ASSIGN_UNMARKED" {
			t.Fatalf("unexpected mutability error: %v", d)
		}
	}
}

func TestMutability_StarMisusedOnRHS_Error(t *testing.T) {
	src := `package p
func f(a int) { *x = *a }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_STAR_MISUSED" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_STAR_MISUSED; diags=%v", res.Diagnostics)
	}
}
