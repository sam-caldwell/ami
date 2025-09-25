package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestContainer_SliceLiteral_AnnotAndInit_OK(t *testing.T) {
	src := `package p
func f() { var s slice<int> = slice<int>{1,2,3} }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_ASSIGN_TYPE_MISMATCH" || d.Code == "E_TYPE_MISMATCH" {
			t.Fatalf("unexpected diag: %v", d)
		}
	}
}

func TestContainer_SetLiteral_Mismatch_Error(t *testing.T) {
	src := `package p
func f() { var s set<int> = set<int>{"a"} }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_MISMATCH" || d.Code == "E_ASSIGN_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a mismatch; diags=%v", res.Diagnostics)
	}
}

func TestContainer_MapLiteral_AnnotAndInit_OK(t *testing.T) {
	src := `package p
func f() { var m map<string,int> = map<string,int>{"a":1, "b":2} }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_MISMATCH" || d.Code == "E_ASSIGN_TYPE_MISMATCH" {
			t.Fatalf("unexpected diag: %v", d)
		}
	}
}

func TestContainer_MapLiteral_KeyTypeMismatch_Error(t *testing.T) {
	src := `package p
func f() { var m map<string,int> = map<string,int>{"a":1, 2:3} }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_TYPE_MISMATCH" || d.Code == "E_ASSIGN_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a mismatch; diags=%v", res.Diagnostics)
	}
}
