package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Positions_Expressions(t *testing.T) {
    src := "package app\nfunc F(){ var x T; x = slice<T>{1}; a.b; Alpha() }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok || fn.Body == nil { t.Fatalf("no func decl or body") }
    if len(fn.Body.Stmts) < 4 { t.Fatalf("expected 4 stmts, got %d", len(fn.Body.Stmts)) }
    // var x T
    if vd, ok := fn.Body.Stmts[0].(*ast.VarDecl); ok {
        if vd.Pos.Line == 0 || vd.NamePos.Line == 0 { t.Fatalf("var positions missing: %+v", vd) }
    } else { t.Fatalf("stmt0 not VarDecl: %T", fn.Body.Stmts[0]) }
    // x = slice<T>{1}
    if as, ok := fn.Body.Stmts[1].(*ast.AssignStmt); ok {
        if as.Pos.Line == 0 || as.NamePos.Line == 0 { t.Fatalf("assign positions missing: %+v", as) }
        if sl, ok := as.Value.(*ast.SliceLit); ok {
            if sl.Pos.Line == 0 || sl.LBrace.Line == 0 || sl.RBrace.Line == 0 { t.Fatalf("slice lit positions missing: %+v", sl) }
        } else { t.Fatalf("assign value not SliceLit: %T", as.Value) }
    } else { t.Fatalf("stmt1 not AssignStmt: %T", fn.Body.Stmts[1]) }
    // a.b
    if es, ok := fn.Body.Stmts[2].(*ast.ExprStmt); ok {
        if sel, ok := es.X.(*ast.SelectorExpr); ok {
            if sel.Pos.Line == 0 || sel.SelPos.Line == 0 { t.Fatalf("selector positions missing: %+v", sel) }
        } else { t.Fatalf("expr not SelectorExpr: %T", es.X) }
    } else { t.Fatalf("stmt2 not ExprStmt: %T", fn.Body.Stmts[2]) }
    // Alpha()
    if es, ok := fn.Body.Stmts[3].(*ast.ExprStmt); ok {
        if ce, ok := es.X.(*ast.CallExpr); ok {
            if ce.Pos.Line == 0 || ce.LParen.Line == 0 || ce.RParen.Line == 0 { t.Fatalf("call positions missing: %+v", ce) }
        } else { t.Fatalf("expr not CallExpr: %T", es.X) }
    } else { t.Fatalf("stmt3 not ExprStmt: %T", fn.Body.Stmts[3]) }
}

