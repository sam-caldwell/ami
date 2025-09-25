package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestAnalyzeEnum_EmptyMembers(t *testing.T) {
	src := `package p
enum X { }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_ENUM_EMPTY" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_ENUM_EMPTY; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEnum_DuplicateMember(t *testing.T) {
	src := `package p
enum X { A, B, A }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_ENUM_DUP_MEMBER" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_ENUM_DUP_MEMBER; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEnum_InvalidValueLiteral(t *testing.T) {
	src := `package p
enum X { A=3.14 }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_ENUM_VALUE_INVALID" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_ENUM_VALUE_INVALID; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEnum_BlankMemberIllegal(t *testing.T) {
	src := `package p
enum X { _ }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_ENUM_BLANK_MEMBER" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_ENUM_BLANK_MEMBER; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEnum_HappyPath(t *testing.T) {
	src := `package p
enum Color { Red, Green, Blue }
enum Code { OK=200, NotFound=404, Custom="X" }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		switch d.Code {
		case "E_ENUM_EMPTY", "E_ENUM_DUP_MEMBER", "E_ENUM_VALUE_INVALID", "E_ENUM_BLANK_MEMBER":
			t.Fatalf("unexpected enum diagnostic: %v", d)
		}
	}
}
