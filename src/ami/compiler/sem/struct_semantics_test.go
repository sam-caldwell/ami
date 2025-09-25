package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestAnalyzeStruct_EmptyFields(t *testing.T) {
	src := `package p
struct S { }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_STRUCT_EMPTY" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_STRUCT_EMPTY; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeStruct_DuplicateField(t *testing.T) {
	src := `package p
struct S { A int, A string }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_STRUCT_DUP_FIELD" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_STRUCT_DUP_FIELD; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeStruct_BlankFieldIllegal(t *testing.T) {
	src := `package p
struct S { _ int }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_STRUCT_BLANK_FIELD" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_STRUCT_BLANK_FIELD; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeStruct_MissingType(t *testing.T) {
	// Omitting field type should yield type invalid diagnostic
	src := `package p
struct S { A }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_STRUCT_FIELD_TYPE_INVALID" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_STRUCT_FIELD_TYPE_INVALID; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeStruct_Happy(t *testing.T) {
	src := `package p
struct Person { Name string, Age int }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		switch d.Code {
		case "E_STRUCT_EMPTY", "E_STRUCT_DUP_FIELD", "E_STRUCT_BLANK_FIELD", "E_STRUCT_FIELD_TYPE_INVALID":
			t.Fatalf("unexpected struct diagnostic: %v", d)
		}
	}
}
