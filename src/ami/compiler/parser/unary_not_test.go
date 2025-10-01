package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_UnaryNotExpression(t *testing.T) {
	code := "package app\nfunc F(){ var x int; !x }\n"
	f := &source.File{Name: "u.ami", Content: code}
	p := New(f)
	af, _ := p.ParseFileCollect()
	if af == nil {
		t.Fatalf("no AST returned")
	}
	// find function F and its body statement
	var found bool
	for _, d := range af.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name == "F" {
			if len(fn.Body.Stmts) == 0 {
				t.Fatalf("no stmts in body")
			}
			if es, ok := fn.Body.Stmts[len(fn.Body.Stmts)-1].(*ast.ExprStmt); ok {
				if _, ok := es.X.(*ast.UnaryExpr); !ok {
					t.Fatalf("expected UnaryExpr, got %T", es.X)
				}
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("function F with unary not not found")
	}
}
