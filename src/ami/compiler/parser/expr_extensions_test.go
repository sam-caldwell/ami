package parser

import (
    "testing"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestParser_Expr_StateSelector_Call(t *testing.T) {
    src := `package p
func f(a int) int {
  state.get(a)
  return a
}`
    p := New(src)
    f := p.ParseFile()
    // find function and first body stmt
    var fn astpkg.FuncDecl
    for _, d := range f.Decls {
        if v, ok := d.(astpkg.FuncDecl); ok { fn = v; break }
    }
    if len(fn.BodyStmts) == 0 {
        t.Fatalf("no body stmts")
    }
    if es, ok := fn.BodyStmts[0].(astpkg.ExprStmt); ok {
        if call, ok2 := es.X.(astpkg.CallExpr); ok2 {
            // expect SelectorExpr fun with X ident 'state'
            if sel, ok3 := call.Fun.(astpkg.SelectorExpr); ok3 {
                if id, ok4 := sel.X.(astpkg.Ident); !ok4 || id.Name != "state" {
                    t.Fatalf("expected receiver 'state', got %#v", sel.X)
                }
            } else {
                t.Fatalf("expected selector fun, got %T", call.Fun)
            }
        } else {
            t.Fatalf("expected call expr, got %T", es.X)
        }
    } else {
        t.Fatalf("expected expr stmt")
    }
}

func TestParser_Expr_GenericLike_Call_TypeArgs(t *testing.T) {
    src := `package p
func f(a int) int {
  Event<uint64>(a)
  return a
}`
    p := New(src)
    f := p.ParseFile()
    var fn astpkg.FuncDecl
    for _, d := range f.Decls {
        if v, ok := d.(astpkg.FuncDecl); ok { fn = v; break }
    }
    if len(fn.BodyStmts) == 0 {
        t.Fatalf("no body stmts")
    }
    if es, ok := fn.BodyStmts[0].(astpkg.ExprStmt); ok {
        if call, ok2 := es.X.(astpkg.CallExpr); ok2 {
            if len(call.TypeArgs) == 0 {
                t.Fatalf("expected type args parsed on call")
            }
        } else {
            t.Fatalf("expected CallExpr, got %T", es.X)
        }
    } else {
        t.Fatalf("expected expr stmt")
    }
}

