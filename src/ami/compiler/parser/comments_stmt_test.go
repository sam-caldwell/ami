package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_FunctionBody_CommentsAttachToStatements(t *testing.T) {
    src := `package app
func F(){
  // before var
  var x T
  /* before assign */
  x = 1
  // before call
  Call()
  // before defer
  defer Done()
  // before return
  return
}`
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl0 not func: %T", file.Decls[0]) }
    if fn.Body == nil { t.Fatalf("no body") }
    // Expect 5 statements
    if len(fn.Body.Stmts) != 5 { t.Fatalf("stmts=%d", len(fn.Body.Stmts)) }
    // var
    if vd, ok := fn.Body.Stmts[0].(*ast.VarDecl); !ok || len(vd.Leading) == 0 || vd.Leading[0].Text == "" {
        t.Fatalf("var leading comments not attached: %#v", fn.Body.Stmts[0])
    }
    // assign
    if as, ok := fn.Body.Stmts[1].(*ast.AssignStmt); !ok || len(as.Leading) == 0 || as.Leading[0].Text == "" {
        t.Fatalf("assign leading comments not attached: %#v", fn.Body.Stmts[1])
    }
    // call expr stmt
    if es, ok := fn.Body.Stmts[2].(*ast.ExprStmt); !ok || len(es.Leading) == 0 || es.Leading[0].Text == "" {
        t.Fatalf("expr leading comments not attached: %#v", fn.Body.Stmts[2])
    }
    // defer
    if ds, ok := fn.Body.Stmts[3].(*ast.DeferStmt); !ok || len(ds.Leading) == 0 || ds.Leading[0].Text == "" {
        t.Fatalf("defer leading comments not attached: %#v", fn.Body.Stmts[3])
    }
    // return
    if rs, ok := fn.Body.Stmts[4].(*ast.ReturnStmt); !ok || len(rs.Leading) == 0 || rs.Leading[0].Text == "" {
        t.Fatalf("return leading comments not attached: %#v", fn.Body.Stmts[4])
    }
}

