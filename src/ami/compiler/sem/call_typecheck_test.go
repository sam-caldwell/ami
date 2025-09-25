package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestCall_ArgTypeMismatch_Error(t *testing.T) {
	src := `package p
func g(a int) {}
func f(a string) { g(a) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_CALL_ARG_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_CALL_ARG_TYPE_MISMATCH; diags=%v", res.Diagnostics)
	}
}

func TestCall_GenericUnify_Event_OK(t *testing.T) {
	src := `package p
func g(ev Event<T>) {}
func f(ev Event<string>) { g(ev) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_CALL_ARG_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %v", d)
		}
	}
}

func TestCall_GenericUnify_Owned_OK(t *testing.T) {
	src := `package p
func h(o Owned<T>) {}
func f(r Owned<string>) { h(r) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_CALL_ARG_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %v", d)
		}
	}
}
