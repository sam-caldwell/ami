package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeCallsWithSigs is like AnalyzeCalls but uses the provided package-wide signature maps.
func AnalyzeCallsWithSigs(f *ast.File, params map[string][]string, results map[string][]string) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    vars := map[string]string{}
    // analyze each function with local var types
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        vars = map[string]string{}
        for _, st := range fn.Body.Stmts {
            if v, ok := st.(*ast.VarDecl); ok {
                if v.Name != "" && v.Type != "" { vars[v.Name] = v.Type }
            }
        }
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.ExprStmt:
                if ce, ok := v.X.(*ast.CallExpr); ok {
                    out = append(out, checkCallWithSigs(ce, params, now, vars)...)
                }
            case *ast.DeferStmt:
                if v.Call != nil { out = append(out, checkCallWithSigs(v.Call, params, now, vars)...)}
            case *ast.ReturnStmt:
                for _, e := range v.Results { if ce, ok := e.(*ast.CallExpr); ok { out = append(out, checkCallWithSigs(ce, params, now, vars)...)} }
            }
        }
    }
    return out
}

func checkCallWithSigs(c *ast.CallExpr, params map[string][]string, now time.Time, vars map[string]string) []diag.Record {
    var out []diag.Record
    if c == nil { return out }
    sigp, ok := params[c.Name]
    if !ok { return out }
    if len(c.Args) != len(sigp) {
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARITY_MISMATCH", Message: "call arity mismatch", Pos: &diag.Position{Line: c.Pos.Line, Column: c.Pos.Column, Offset: c.Pos.Offset}})
        return out
    }
    for i, a := range c.Args {
        at := inferExprTypeWithVars(a, vars)
        pt := sigp[i]
        if pt == "" || pt == "any" || at == "any" { continue }
        if pt != at {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARG_TYPE_MISMATCH", Message: "call argument type mismatch", Pos: &diag.Position{Line: c.Pos.Line, Column: c.Pos.Column, Offset: c.Pos.Offset}})
            break
        }
    }
    return out
}

