package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testAnalyzeEnums_DuplicateAndBlank(t *testing.T) {
	code := "package app\nenum X { A, , A }\n"
	f := (&source.FileSet{}).AddFile("e1.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeEnums(af)
	var hasDup, hasBlank bool
	for _, d := range ds {
		if d.Code == "E_ENUM_MEMBER_DUP" {
			hasDup = true
		}
		if d.Code == "E_ENUM_MEMBER_BLANK" {
			hasBlank = true
		}
	}
	if !hasDup || !hasBlank {
		t.Fatalf("expected dup+blank: %+v", ds)
	}
}

func testAnalyzeEnums_AssignmentValidation(t *testing.T) {
	code := "package app\nenum Color { Red, Green }\nfunc F(){ var x Color = Red; var y Color = 123 }\n"
	f := (&source.FileSet{}).AddFile("e2.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeEnums(af)
	var hasInvalid bool
	for _, d := range ds {
		if d.Code == "E_ENUM_ASSIGN_INVALID" {
			hasInvalid = true
		}
	}
	if !hasInvalid {
		t.Fatalf("expected E_ENUM_ASSIGN_INVALID: %+v", ds)
	}
}
