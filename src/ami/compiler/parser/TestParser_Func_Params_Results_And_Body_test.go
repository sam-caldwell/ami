package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Func_Params_Results_And_Body(t *testing.T) {
    // typed params/returns and statements: var, call, return
    src := "package app\nfunc F(a T, b U) (R1,R2) { var x T; Alpha(); return a,b }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl is %T", file.Decls[0]) }
    if len(fn.Results) != 2 { t.Fatalf("want 2 results, got %d", len(fn.Results)) }
    // Find the return statement and validate tuple arity
    if fn.Body == nil { t.Fatalf("no body") }
    var saw bool
    for _, st := range fn.Body.Stmts {
        if rs, ok := st.(*ast.ReturnStmt); ok {
            if len(rs.Results) != 2 { t.Fatalf("want 2 return exprs, got %d", len(rs.Results)) }
            saw = true
        }
    }
    if !saw { t.Fatalf("no return found") }
}

