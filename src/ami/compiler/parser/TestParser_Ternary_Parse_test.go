package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParser_Ternary_Parse ensures the parser recognizes the C-style ternary operator.
func TestParser_Ternary_Parse(t *testing.T) {
	code := "package app\nfunc F(){ var a int; var b int; (a == 1) ? b : 2 }\n"
	f := &source.File{Name: "tern.ami", Content: code}
	p := New(f)
	af, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	found := false
	for _, d := range af.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name != "F" || fn.Body == nil {
			continue
		}
		for _, st := range fn.Body.Stmts {
			if es, ok := st.(*ast.ExprStmt); ok {
				if _, ok := es.X.(*ast.ConditionalExpr); ok {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected to find ConditionalExpr in return statement")
	}
}
