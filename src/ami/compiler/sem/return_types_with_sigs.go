package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeReturnTypesWithSigs extends AnalyzeReturnTypes by using known function results
// to infer call expression types and detect E_CALL_RESULT_MISMATCH at return sites.
func AnalyzeReturnTypesWithSigs(f *ast.File, results map[string][]string) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        decl := make([]string, len(fn.Results))
        for i, r := range fn.Results { decl[i] = r.Type }
        // Build a simple local type env from params, var decls, and assignments.
        env := buildLocalEnv(fn)
        for _, st := range fn.Body.Stmts {
            rs, ok := st.(*ast.ReturnStmt)
            if !ok { continue }
            // infer result types
            var got []string
            if len(rs.Results) == 1 {
                // Special case: single call expression returning multiple results.
                if ce, ok := rs.Results[0].(*ast.CallExpr); ok {
                    if rts, ok := results[ce.Name]; ok && len(rts) > 1 {
                        got = append(got, rts...)
                    }
                }
            }
            if len(got) == 0 {
                // Mixed expressions: expand the first multi-result call when it fits arity.
                expanded := false
                // compute declared arity
                declN := len(decl)
                // count non-call exprs
                nonCall := 0
                for _, e := range rs.Results { if _, isCall := e.(*ast.CallExpr); !isCall { nonCall++ } }
                for i, e := range rs.Results {
                    if ce, isCall := e.(*ast.CallExpr); isCall && !expanded {
                        if rts, ok := results[ce.Name]; ok && nonCall+len(rts) == declN {
                            got = append(got, rts...)
                            expanded = true
                            continue
                        }
                    }
                    got = append(got, inferRetTypeWithEnv(e, env, results))
                    // after first element, mark expanded if we already appended rts
                    _ = i
                }
            }
            if len(got) != len(decl) {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return arity mismatch", Pos: &diag.Position{Line: rs.Pos.Line, Column: rs.Pos.Column, Offset: rs.Pos.Offset}})
                continue
            }
            // compare
            mismatch := false
            for i := range got {
                if decl[i] == "" || decl[i] == "any" || got[i] == "any" { continue }
                if decl[i] != got[i] { mismatch = true; break }
            }
            if mismatch {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_RESULT_MISMATCH", Message: "return type mismatch from call", Pos: &diag.Position{Line: rs.Pos.Line, Column: rs.Pos.Column, Offset: rs.Pos.Offset}})
            }
        }
    }
    return out
}

func inferRetTypeWithEnv(e ast.Expr, env map[string]string, results map[string][]string) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if t := env[v.Name]; t != "" { return t }
        return "any"
    case *ast.CallExpr:
        if rs, ok := results[v.Name]; ok && len(rs) > 0 && rs[0] != "" { return rs[0] }
        return "any"
    case *ast.NumberLit, *ast.StringLit, *ast.SliceLit, *ast.SetLit, *ast.MapLit:
        return deduceType(e)
    default:
        return inferExprType(env, e)
    }
}

// buildLocalEnv collects local variable types from parameters, var declarations,
// and simple assignments where the right-hand side has a deducible type.
func buildLocalEnv(fn *ast.FuncDecl) map[string]string {
    env := map[string]string{}
    for _, p := range fn.Params { if p.Name != "" && p.Type != "" { env[p.Name] = p.Type } }
    if fn.Body == nil { return env }
    for _, st := range fn.Body.Stmts {
        switch v := st.(type) {
        case *ast.VarDecl:
            if v.Name != "" {
                if v.Type != "" { env[v.Name] = v.Type } else if v.Init != nil {
                    if t := deduceType(v.Init); t != "any" && t != "" { env[v.Name] = t }
                }
            }
        case *ast.AssignStmt:
            if v.Name != "" && v.Value != nil {
                if t := deduceType(v.Value); t != "any" && t != "" { env[v.Name] = t }
            }
        }
    }
    return env
}
