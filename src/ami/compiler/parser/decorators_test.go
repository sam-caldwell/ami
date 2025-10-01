package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_Decorators_Single_NoArgs(t *testing.T) {
	src := "package p\n@trace\nfunc F() {}\n"
	var fs source.FileSet
	f := fs.AddFile("f.ami", src)
	p := New(f)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(file.Decls))
	}
	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("decl 0 not a func")
	}
	if len(fn.Decorators) != 1 {
		t.Fatalf("expected 1 decorator, got %d", len(fn.Decorators))
	}
	if fn.Decorators[0].Name != "trace" {
		t.Fatalf("decorator name: %s", fn.Decorators[0].Name)
	}
}

func TestParse_Decorators_Multiple_WithArgs_OrderPreserved(t *testing.T) {
	src := "package p\n@deprecated(\"x\")\n@metrics(1, id)\nfunc F() {}\n"
	var fs source.FileSet
	f := fs.AddFile("f.ami", src)
	p := New(f)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(file.Decls))
	}
	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("decl 0 not a func")
	}
	if got := len(fn.Decorators); got != 2 {
		t.Fatalf("decorators: %d", got)
	}
	if fn.Decorators[0].Name != "deprecated" {
		t.Fatalf("first: %s", fn.Decorators[0].Name)
	}
	if fn.Decorators[1].Name != "metrics" {
		t.Fatalf("second: %s", fn.Decorators[1].Name)
	}
	if len(fn.Decorators[1].Args) != 2 {
		t.Fatalf("metrics args: %d", len(fn.Decorators[1].Args))
	}
	// first arg should be number literal "1"
	if num, ok := fn.Decorators[1].Args[0].(*ast.NumberLit); !ok || num.Text != "1" {
		t.Fatalf("arg0 not number '1': %#v", fn.Decorators[1].Args[0])
	}
	// second arg should be identifier 'id'
	if id, ok := fn.Decorators[1].Args[1].(*ast.IdentExpr); !ok || id.Name != "id" {
		t.Fatalf("arg1 not ident 'id': %#v", fn.Decorators[1].Args[1])
	}
}
