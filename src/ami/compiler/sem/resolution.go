package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeNameResolution performs simple intra-function name resolution for identifiers
// used in expressions. Emits E_UNRESOLVED_IDENT with source positions when an
// identifier is used without a corresponding parameter or local declaration.
func AnalyzeNameResolution(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // collect top-level function names to allow referencing them as values (e.g., passing handlers)
    topFuncs := map[string]struct{}{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok && fn.Name != "" {
            topFuncs[fn.Name] = struct{}{}
        }
    }
    // collect import aliases for scope: alias or last path segment
    imports := map[string]struct{}{}
    for _, d := range f.Decls {
        if im, ok := d.(*ast.ImportDecl); ok {
            alias := im.Alias
            if alias == "" {
                // derive from last path segment
                p := im.Path
                if i := lastSlash(p); i >= 0 && i+1 < len(p) { alias = p[i+1:] } else { alias = p }
            }
            if alias != "" { imports[alias] = struct{}{} }
        }
    }
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        env := map[string]bool{}
        for _, p := range fn.Params { if p.Name != "" { env[p.Name] = true } }
        // allow referencing any top-level func names in this file
        for n := range topFuncs { env[n] = true }
        // Gather var decls
        for _, st := range fn.Body.Stmts {
            if vd, ok := st.(*ast.VarDecl); ok && vd.Name != "" { env[vd.Name] = true }
        }
        // Walk expressions to find unresolved idents
        var walkExpr func(e ast.Expr)
        walkExpr = func(e ast.Expr) {
            switch v := e.(type) {
            case *ast.IdentExpr:
                if v.Name != "" && !env[v.Name] {
                    if _, ok := imports[v.Name]; ok { return }
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_UNRESOLVED_IDENT", Message: "unresolved identifier: " + v.Name, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
                }
            case *ast.BinaryExpr:
                walkExpr(v.X); walkExpr(v.Y)
            case *ast.CallExpr:
                for _, a := range v.Args { walkExpr(a) }
            case *ast.SliceLit:
                for _, a := range v.Elems { walkExpr(a) }
            case *ast.SetLit:
                for _, a := range v.Elems { walkExpr(a) }
            case *ast.MapLit:
                for _, kv := range v.Elems { walkExpr(kv.Key); walkExpr(kv.Val) }
            }
        }
        // Apply over statements
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.AssignStmt:
                walkExpr(v.Value)
            case *ast.ReturnStmt:
                for _, e := range v.Results { walkExpr(e) }
            case *ast.ExprStmt:
                walkExpr(v.X)
            case *ast.VarDecl:
                if v.Init != nil { walkExpr(v.Init) }
            }
        }
    }
    return out
}

func lastSlash(s string) int {
    for i := len(s) - 1; i >= 0; i-- { if s[i] == '/' { return i } }
    return -1
}
