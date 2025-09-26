package parser

import (
	ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestParser_Body_Positions_DeferAndCall(t *testing.T) {
	src := `package p
func f(ev Event<string>) (Event<string>, error) {
    defer cleanup()
    x()
}`
	p := New(src)
	f := p.ParseFile()
	fd := f.Decls[0].(ast.FuncDecl)
	if len(fd.BodyStmts) != 2 {
		t.Fatalf("expected 2 body stmts; got %d", len(fd.BodyStmts))
	}
	if ds, ok := fd.BodyStmts[0].(ast.DeferStmt); ok {
		if ds.Pos.Line <= 0 {
			t.Fatalf("defer stmt pos not set")
		}
		if call, ok := ds.X.(ast.CallExpr); ok {
			if call.Pos.Line <= 0 {
				t.Fatalf("call expr pos not set within defer")
			}
		} else {
			t.Fatalf("expected call in defer; got %T", ds.X)
		}
	} else {
		t.Fatalf("expected DeferStmt; got %T", fd.BodyStmts[0])
	}
	if es, ok := fd.BodyStmts[1].(ast.ExprStmt); ok {
		if es.Pos.Line <= 0 {
			t.Fatalf("expr stmt pos not set")
		}
	} else {
		t.Fatalf("expected ExprStmt; got %T", fd.BodyStmts[1])
	}
}
