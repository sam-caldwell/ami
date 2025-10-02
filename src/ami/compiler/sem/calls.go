package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeCalls checks call sites against known function signatures to detect arity
// and basic argument type mismatches (scaffold typing rules).
func AnalyzeCalls(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // collect function signatures
    funcs := map[string]sig{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            var ps []string
            var rs []string
            var pnames []string
            var ppos []diag.Position
            for _, p := range fn.Params {
                ps = append(ps, p.Type)
                pnames = append(pnames, p.Name)
                tp := diag.Position{Line: p.TypePos.Line, Column: p.TypePos.Column, Offset: p.TypePos.Offset}
                if p.TypePos.Line == 0 { tp = diag.Position{Line: p.Pos.Line, Column: p.Pos.Column, Offset: p.Pos.Offset} }
                ppos = append(ppos, tp)
            }
            for _, r := range fn.Results { rs = append(rs, r.Type) }
            funcs[fn.Name] = sig{params: ps, results: rs, paramNames: pnames, paramTypePos: ppos}
        }
    }
    // analyze each function body
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        // collect local var types from params, var decls/inits and assignments
        vars := map[string]string{}
        for _, p := range fn.Params { if p.Name != "" && p.Type != "" { vars[p.Name] = p.Type } }
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.VarDecl:
                if v.Name != "" {
                    if v.Type != "" { vars[v.Name] = v.Type } else if v.Init != nil {
                        if t := deduceType(v.Init); t != "any" && t != "" { vars[v.Name] = t }
                    }
                }
            case *ast.AssignStmt:
                if v.Name != "" && v.Value != nil {
                    if t := deduceType(v.Value); t != "any" && t != "" { vars[v.Name] = t }
                }
            }
        }
        // walk statements for calls and return calls
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.ExprStmt:
                if ce, ok := v.X.(*ast.CallExpr); ok { out = append(out, checkCall(ce, funcs, vars, now)...) }
            case *ast.DeferStmt:
                if v.Call != nil { out = append(out, checkCall(v.Call, funcs, vars, now)...) }
            case *ast.ReturnStmt:
                for _, e := range v.Results {
                    if ce, ok := e.(*ast.CallExpr); ok { out = append(out, checkCall(ce, funcs, vars, now)...) }
                }
            }
        }
    }
    return out
}
