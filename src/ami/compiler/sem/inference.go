package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeTypeInference performs M1 local type inference for identifiers and
// unary/binary expressions inside function bodies, emitting diagnostics with
// precise positions on mismatches and unknowns.
func AnalyzeTypeInference(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        // build local env
        env := map[string]string{}
        for _, p := range fn.Params {
            if p.Name != "" && p.Type != "" { env[p.Name] = p.Type }
        }
        // collect local function result signatures for call propagation
        sigs := collectFunctionResults(f)
        // seed from var decls
        for _, st := range fn.Body.Stmts {
            if vd, ok := st.(*ast.VarDecl); ok {
                if vd.Name != "" {
                    if vd.Type != "" {
                        env[vd.Name] = vd.Type
                    } else if vd.Init != nil {
                        env[vd.Name] = inferLocalExprTypeWithSigs(env, sigs, vd.Init)
                    }
                }
            }
        }
        // conservative multi-pass propagation over assignments to reach a fixed point quickly
        for pass := 0; pass < 3; pass++ {
            changed := false
            for _, st := range fn.Body.Stmts {
                as, ok := st.(*ast.AssignStmt)
                if !ok { continue }
                vt := inferLocalExprTypeWithSigs(env, sigs, as.Value)
                if vt == "" || vt == "any" { continue }
                old := env[as.Name]
                if old == "" || old == "any" {
                    env[as.Name] = vt
                    changed = true
                    continue
                }
                // if both known but container generics unify, keep the old to avoid churn
                // otherwise leave as-is; mismatch will be reported below
            }
            if !changed { break }
        }
        // diagnostics: type mismatches and ambiguous returns
        for _, st := range fn.Body.Stmts {
            switch v := st.(type) {
            case *ast.AssignStmt:
                vt := inferLocalExprTypeWithSigs(env, sigs, v.Value)
                if old, ok := env[v.Name]; ok && old != "" {
                    if !typesCompatible(old, vt) {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "assignment type mismatch: expected " + old + ", got " + vt, Pos: &diag.Position{Line: v.NamePos.Line, Column: v.NamePos.Column, Offset: v.NamePos.Offset}})
                    }
                }
            case *ast.ReturnStmt:
                for _, e := range v.Results { out = append(out, ambiguousInExpr(now, e)...)}
            }
        }
    }
    return out
}
