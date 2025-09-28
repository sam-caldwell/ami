package sem

import (
    "fmt"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

type sig struct{ params, results []string; paramNames []string }

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
            for _, p := range fn.Params { ps = append(ps, p.Type); pnames = append(pnames, p.Name) }
            for _, r := range fn.Results { rs = append(rs, r.Type) }
            funcs[fn.Name] = sig{params: ps, results: rs, paramNames: pnames}
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

func checkCall(c *ast.CallExpr, funcs map[string]sig, vars map[string]string, now time.Time) []diag.Record {
    var out []diag.Record
    if c == nil { return out }
    s, ok := funcs[c.Name]
    if !ok { return out }
    // arity
    if len(c.Args) != len(s.params) {
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARITY_MISMATCH", Message: "call arity mismatch", Pos: &diag.Position{Line: c.Pos.Line, Column: c.Pos.Column, Offset: c.Pos.Offset}})
        return out
    }
    // type checks
    for i, a := range c.Args {
        at := inferExprTypeWithVars(a, vars)
        pt := s.params[i]
        if pt == "" || pt == "any" || at == "any" { continue }
        if !typesCompatible(pt, at) {
            p := epos(a)
            msg := fmt.Sprintf("call argument type mismatch: arg %d expected %s, got %s", i, pt, at)
            data := map[string]any{"argIndex": i, "expected": pt, "actual": at, "callee": c.Name}
            if i < len(s.paramNames) && s.paramNames[i] != "" { data["paramName"] = s.paramNames[i] }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARG_TYPE_MISMATCH", Message: msg, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: data})
        }
    }
    return out
}

func inferExprTypeWithVars(e ast.Expr, vars map[string]string) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if t, ok := vars[v.Name]; ok && t != "" { return t }
        return "any"
    case *ast.StringLit:
        return "string"
    case *ast.NumberLit:
        return "int"
    case *ast.SliceLit, *ast.SetLit, *ast.MapLit:
        return deduceType(e)
    case *ast.CallExpr:
        // unknown without inter-procedural lookup here
        return "any"
    default:
        return "any"
    }
}
