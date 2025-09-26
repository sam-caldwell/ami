package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "testing"
)

// Ensure statements without leading comments have empty Comments slices.
func TestParser_BodyComments_NoCommentsAreEmpty(t *testing.T) {
    src := `package p
func f() {
  // have comment
  var x = 1
  a + 2
  a = 3
  defer f()
  return
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
        t.Fatalf("expected at least 5 body statements, got %d", len(fn.BodyStmts))
    }
    // VarDecl has comment (control)
    if v, ok := fn.BodyStmts[0].(astpkg.VarDeclStmt); ok {
        if len(v.Comments) == 0 {
            t.Fatalf("expected a comment on var decl (control)")
        }
    } else {
        t.Fatalf("expected VarDeclStmt first")
    }
    // The rest should have no comments
    if es, ok := fn.BodyStmts[1].(astpkg.ExprStmt); ok {
        if len(es.Comments) != 0 {
            t.Fatalf("expected no comments on expr stmt; got %+v", es.Comments)
        }
    } else { t.Fatalf("expected ExprStmt second") }

    if as, ok := fn.BodyStmts[2].(astpkg.AssignStmt); ok {
        if len(as.Comments) != 0 {
            t.Fatalf("expected no comments on assign stmt; got %+v", as.Comments)
        }
    } else { t.Fatalf("expected AssignStmt third") }

    if ds, ok := fn.BodyStmts[3].(astpkg.DeferStmt); ok {
        if len(ds.Comments) != 0 {
            t.Fatalf("expected no comments on defer stmt; got %+v", ds.Comments)
        }
    } else { t.Fatalf("expected DeferStmt fourth") }

    if rs, ok := fn.BodyStmts[4].(astpkg.ReturnStmt); ok {
        if len(rs.Comments) != 0 {
            t.Fatalf("expected no comments on return stmt; got %+v", rs.Comments)
        }
    } else { t.Fatalf("expected ReturnStmt fifth") }
}

