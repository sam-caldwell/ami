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
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        env := map[string]bool{}
        for _, p := range fn.Params { if p.Name != "" { env[p.Name] = true } }
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

