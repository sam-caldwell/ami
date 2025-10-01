package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestParser_UnaryMinus_And_BitwiseNot(t *testing.T) {
	code := "package app\nfunc F(){ var x int; -x; ~x }\n"
	f := &source.File{Name: "u2.ami", Content: code}
	p := New(f)
	af, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var foundMinus, foundBnot bool
	for _, d := range af.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name != "F" {
			continue
		}
		for _, st := range fn.Body.Stmts {
			if es, ok := st.(*ast.ExprStmt); ok {
				if ue, ok := es.X.(*ast.UnaryExpr); ok {
					if ue.Op == token.Minus {
						foundMinus = true
					}
					if ue.Op == token.TildeSym {
						foundBnot = true
					}
				}
			}
		}
	}
	if !foundMinus || !foundBnot {
		t.Fatalf("missing unary ops: -=%v ~= %v", foundMinus, foundBnot)
	}
}
