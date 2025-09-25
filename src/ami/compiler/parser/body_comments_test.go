package parser

import (
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestParser_BodyComments_AttachedToStatements(t *testing.T) {
	src := `package p
func f(a int) (int,int) {
  // before var
  var x = a + 1
  // before expr
  a + 2
  // before assign
  *a = a + 3
  // before defer
  defer f(a)
  // before return
  return a, a
}`
	p := New(src)
	file := p.ParseFile()
	// find function decl
	var fn astpkg.FuncDecl
	for _, d := range file.Decls {
		if v, ok := d.(astpkg.FuncDecl); ok {
			fn = v
			break
		}
	}
	if fn.Name == "" {
		t.Fatalf("expected func decl present")
	}
	if len(fn.BodyStmts) < 5 {
		t.Fatalf("expected body statements, got %d", len(fn.BodyStmts))
	}
	// check VarDecl comment
	if v, ok := fn.BodyStmts[0].(astpkg.VarDeclStmt); ok {
		if len(v.Comments) == 0 || v.Comments[0].Text == "" {
			t.Fatalf("missing comment on var decl: %+v", v.Comments)
		}
	} else {
		t.Fatalf("expected VarDeclStmt first")
	}
	// check ExprStmt comment
	if es, ok := fn.BodyStmts[1].(astpkg.ExprStmt); ok {
		if len(es.Comments) == 0 || es.Comments[0].Text == "" {
			t.Fatalf("missing comment on expr stmt: %+v", es.Comments)
		}
	} else {
		t.Fatalf("expected ExprStmt second")
	}
	// check AssignStmt comment
	if as, ok := fn.BodyStmts[2].(astpkg.AssignStmt); ok {
		if len(as.Comments) == 0 || as.Comments[0].Text == "" {
			t.Fatalf("missing comment on assign stmt: %+v", as.Comments)
		}
	} else {
		t.Fatalf("expected AssignStmt third")
	}
	// check DeferStmt comment
	if ds, ok := fn.BodyStmts[3].(astpkg.DeferStmt); ok {
		if len(ds.Comments) == 0 || ds.Comments[0].Text == "" {
			t.Fatalf("missing comment on defer stmt: %+v", ds.Comments)
		}
	} else {
		t.Fatalf("expected DeferStmt fourth")
	}
	// check ReturnStmt comment
	if rs, ok := fn.BodyStmts[4].(astpkg.ReturnStmt); ok {
		if len(rs.Comments) == 0 || rs.Comments[0].Text == "" {
			t.Fatalf("missing comment on return stmt: %+v", rs.Comments)
		}
	} else {
		t.Fatalf("expected ReturnStmt fifth")
	}
}
