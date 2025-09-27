package sem

import (
    "fmt"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeCallsWithSigs is like AnalyzeCalls but uses the provided package-wide signature maps.
func AnalyzeCallsWithSigs(f *ast.File, params map[string][]string, results map[string][]string) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // analyze each function with local var types
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        // build local env (params, var inits/types, assignments)
        vars := buildLocalEnv(fn)
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
            p := epos(a)
            msg := fmt.Sprintf("call argument type mismatch: arg %d expected %s, got %s", i, pt, at)
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARG_TYPE_MISMATCH", Message: msg, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: map[string]any{"argIndex": i, "expected": pt, "actual": at, "callee": c.Name}})
        }
    }
    return out
}
