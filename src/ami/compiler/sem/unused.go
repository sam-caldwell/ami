package sem

import (
    "time"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeUnused emits warnings for unused locals and functions within a file.
// Diagnostics:
//  - W_UNUSED_VAR: variable declared but never used
//  - W_UNUSED_FUNC: function declared but never referenced
func AnalyzeUnused(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // Collect locals per function and mark uses
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        locals := map[string]diag.Position{}
        used := map[string]bool{}
        for _, p := range fn.Params {
            if p.Name != "" { locals[p.Name] = diag.Position{Line: p.Pos.Line, Column: p.Pos.Column, Offset: p.Pos.Offset} }
        }
        for _, st := range fn.Body.Stmts {
            if vd, ok := st.(*ast.VarDecl); ok && vd.Name != "" {
                locals[vd.Name] = diag.Position{Line: vd.Pos.Line, Column: vd.Pos.Column, Offset: vd.Pos.Offset}
            }
        }
        // Walk body to mark identifier uses
        walkStmts := func(st ast.Stmt) {}
        var walkExpr func(e ast.Expr)
        walkExpr = func(e ast.Expr) {
            switch v := e.(type) {
            case *ast.IdentExpr:
                used[v.Name] = true
            case *ast.CallExpr:
                for _, a := range v.Args { walkExpr(a) }
            case *ast.UnaryExpr:
                walkExpr(v.X)
            case *ast.BinaryExpr:
                walkExpr(v.X); walkExpr(v.Y)
            case *ast.SliceLit:
                for _, el := range v.Elems { walkExpr(el) }
            case *ast.SetLit:
                for _, el := range v.Elems { walkExpr(el) }
            case *ast.MapLit:
                for _, kv := range v.Elems { walkExpr(kv.Key); walkExpr(kv.Val) }
            }
        }
        walkStmts = func(st ast.Stmt) {
            switch v := st.(type) {
            case *ast.ExprStmt:
                if v.X != nil { walkExpr(v.X) }
            case *ast.AssignStmt:
                if v.Value != nil { walkExpr(v.Value) }
            case *ast.VarDecl:
                if v.Init != nil { walkExpr(v.Init) }
            case *ast.ReturnStmt:
                for _, e := range v.Results { walkExpr(e) }
            case *ast.DeferStmt:
                if v.Call != nil { walkExpr(v.Call) }
            }
        }
        for _, st := range fn.Body.Stmts { walkStmts(st) }
        for name, pos := range locals {
            if used[name] { continue }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_UNUSED_VAR", Message: "unused variable: " + name, Pos: &pos})
        }
    }
    // Unused functions: collect definitions and references
    defs := map[string]diag.Position{}
    refs := map[string]bool{}
    for _, d := range f.Decls { if fn, ok := d.(*ast.FuncDecl); ok { defs[fn.Name] = diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset} } }
    // mark calls
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
            for _, st := range fn.Body.Stmts {
                if es, ok := st.(*ast.ExprStmt); ok {
                    if ce, ok := es.X.(*ast.CallExpr); ok { refs[ce.Name] = true }
                }
                if ds, ok := st.(*ast.DeferStmt); ok && ds.Call != nil { refs[ds.Call.Name] = true }
                if rs, ok := st.(*ast.ReturnStmt); ok {
                    for _, e := range rs.Results { if ce, ok := e.(*ast.CallExpr); ok { refs[ce.Name] = true } }
                }
            }
        }
    }
    // root entrypoint: main is considered used
    refs["main"] = true
    for name, p := range defs {
        if refs[name] { continue }
        pos := p
        out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_UNUSED_FUNC", Message: "unused function: " + name, Pos: &pos})
    }
    return out
}
