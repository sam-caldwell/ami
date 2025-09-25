package parser

import (
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestParser_Import_Unquoted_WithConstraint_Single(t *testing.T) {
	src := "package p\nimport ami/stdlib/io >= v0.1.2\n"
	p := New(src)
	f := p.ParseFile()
	// find ImportDecl
	found := false
	for _, d := range f.Decls {
		if id, ok := d.(astpkg.ImportDecl); ok {
			if id.Path == "ami/stdlib/io" && id.Constraint == ">= v0.1.2" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("import with constraint not found: %+v", f.Decls)
	}
	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parse errors: %+v", p.Errors())
	}
}

func TestParser_Import_Unquoted_WithConstraint_Block(t *testing.T) {
	src := "package p\nimport (\n  ami/stdlib/io >= v0.0.1\n  github.com/asymmetric-effort/ami/stdio >= v0.0.0\n)\n"
	p := New(src)
	f := p.ParseFile()
	got := map[string]string{}
	for _, d := range f.Decls {
		if id, ok := d.(astpkg.ImportDecl); ok {
			got[id.Path] = id.Constraint
		}
	}
	if got["ami/stdlib/io"] != ">= v0.0.1" {
		t.Fatalf("unexpected io constraint: %q", got["ami/stdlib/io"])
	}
	if got["github.com/asymmetric-effort/ami/stdio"] != ">= v0.0.0" {
		t.Fatalf("unexpected stdio constraint: %q", got["github.com/asymmetric-effort/ami/stdio"])
	}
}
