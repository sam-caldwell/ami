package sem

import (
    "fmt"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeCallsWithSigs is like AnalyzeCalls but uses the provided package-wide signature maps,
// and optionally parameter type positions from the driver. When positions are not provided,
// it falls back to local function declarations in the same file.
func AnalyzeCallsWithSigs(f *ast.File, params map[string][]string, results map[string][]string, paramPos map[string][]diag.Position, paramNames map[string][]string) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // Gather local param type positions for functions present in this file (best-effort)
    localParamPos := map[string][]diag.Position{}
    localParamNames := map[string][]string{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            var pos []diag.Position
            var names []string
            for _, p := range fn.Params {
                tp := diag.Position{Line: p.TypePos.Line, Column: p.TypePos.Column, Offset: p.TypePos.Offset}
                if p.TypePos.Line == 0 { tp = diag.Position{Line: p.Pos.Line, Column: p.Pos.Column, Offset: p.Pos.Offset} }
                pos = append(pos, tp)
                names = append(names, p.Name)
            }
            localParamPos[fn.Name] = pos
            localParamNames[fn.Name] = names
        }
    }
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
                    // prefer driver-provided positions when present
                    effective := localParamPos
                    if paramPos != nil { effective = paramPos }
                    effNames := localParamNames
                    if paramNames != nil { effNames = paramNames }
                    out = append(out, checkCallWithSigsWithResults(ce, params, results, now, vars, effective, effNames)...)
                }
            case *ast.DeferStmt:
                if v.Call != nil {
                    effective := localParamPos
                    if paramPos != nil { effective = paramPos }
                    effNames := localParamNames
                    if paramNames != nil { effNames = paramNames }
                    out = append(out, checkCallWithSigsWithResults(v.Call, params, results, now, vars, effective, effNames)...)
                }
            case *ast.ReturnStmt:
                for _, e := range v.Results {
                    if ce, ok := e.(*ast.CallExpr); ok {
                        effective := localParamPos
                        if paramPos != nil { effective = paramPos }
                        effNames := localParamNames
                        if paramNames != nil { effNames = paramNames }
                        out = append(out, checkCallWithSigsWithResults(ce, params, results, now, vars, effective, effNames)...)
                    }
                }
            }
        }
    }
    return out
}

func checkCallWithSigsWithResults(c *ast.CallExpr, params map[string][]string, results map[string][]string, now time.Time, vars map[string]string, paramPos map[string][]diag.Position, paramNames map[string][]string) []diag.Record {
    var out []diag.Record
    if c == nil { return out }
    sigp, ok := params[c.Name]
    if !ok { return out }
    if len(c.Args) != len(sigp) {
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARITY_MISMATCH", Message: "call arity mismatch", Pos: &diag.Position{Line: c.Pos.Line, Column: c.Pos.Column, Offset: c.Pos.Offset}})
        return out
    }
    var mismatchIdx []int
    for i, a := range c.Args {
        at := inferExprTypeWithEnvAndResults(a, vars, results)
        pt := sigp[i]
        if pt == "" || pt == "any" || at == "any" { continue }
        if mismatch, path, pathIdx, fieldPath, base, wantN, gotN := findGenericArityMismatchWithFields(pt, at); mismatch {
            p := epos(a)
            data := map[string]any{"argIndex": i, "callee": c.Name, "base": base, "path": path, "pathIdx": pathIdx, "fieldPath": fieldPath, "expected": pt, "actual": at, "expectedArity": wantN, "actualArity": gotN}
            if v, ok := paramPos[c.Name]; ok && i < len(v) { data["expectedPos"] = v[i] }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_GENERIC_ARITY_MISMATCH", Message: "generic type argument count mismatch", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: data})
            mismatchIdx = append(mismatchIdx, i)
            continue
        } else if mismatch, path, pathIdx, fieldPath, base, wantN, gotN := findGenericArityMismatchDeepPathTextFields(pt, at); mismatch {
            p := epos(a)
            data := map[string]any{"argIndex": i, "callee": c.Name, "base": base, "path": path, "pathIdx": pathIdx, "expected": pt, "actual": at, "expectedArity": wantN, "actualArity": gotN}
            if len(fieldPath) > 0 { data["fieldPath"] = fieldPath }
            if v, ok := paramPos[c.Name]; ok && i < len(v) { data["expectedPos"] = v[i] }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_GENERIC_ARITY_MISMATCH", Message: "generic type argument count mismatch", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: data})
            mismatchIdx = append(mismatchIdx, i)
            continue
        }
        if !typesCompatible(pt, at) {
            p := epos(a)
            msg := fmt.Sprintf("call argument type mismatch: arg %d expected %s, got %s", i, pt, at)
            data := map[string]any{"argIndex": i, "expected": pt, "actual": at, "callee": c.Name}
            if v, ok := paramPos[c.Name]; ok && i < len(v) { data["expectedPos"] = v[i] }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARG_TYPE_MISMATCH", Message: msg, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: data})
            mismatchIdx = append(mismatchIdx, i)
        }
    }
    if len(mismatchIdx) > 1 {
        var paths []map[string]any
        for _, i := range mismatchIdx {
            pt := sigp[i]
            at := inferExprTypeWithEnvAndResults(c.Args[i], vars, results)
            if m, p, idx, fp, b, _, _ := findGenericArityMismatchWithFields(pt, at); m {
                e := map[string]any{"argIndex": i, "base": b, "path": p, "pathIdx": idx, "fieldPath": fp}
                if v, ok := paramPos[c.Name]; ok && i < len(v) { e["expectedPos"] = v[i] }
                if v, ok := paramNames[c.Name]; ok && i < len(v) && v[i] != "" { e["paramName"] = v[i] }
                paths = append(paths, e)
            } else if m2, p2, idx2, fp2, b2, _, _ := findGenericArityMismatchDeepPathTextFields(pt, at); m2 {
                e := map[string]any{"argIndex": i, "base": b2, "path": p2, "pathIdx": idx2}
                if len(fp2) > 0 { e["fieldPath"] = fp2 }
                if v, ok := paramPos[c.Name]; ok && i < len(v) { e["expectedPos"] = v[i] }
                if v, ok := paramNames[c.Name]; ok && i < len(v) && v[i] != "" { e["paramName"] = v[i] }
                paths = append(paths, e)
            }
        }
        data := map[string]any{"count": len(mismatchIdx), "indices": mismatchIdx, "callee": c.Name}
        if len(paths) > 0 { data["paths"] = paths }
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARGS_MISMATCH_SUMMARY", Message: "multiple call arguments mismatch", Pos: &diag.Position{Line: c.Pos.Line, Column: c.Pos.Column, Offset: c.Pos.Offset}, Data: data})
    }
    return out
}

// paramsToResults adapts a params map to a results‑like map (not used now).
func paramsToResults(params map[string][]string) map[string][]string { return params }

// inferExprTypeWithEnvAndResults attempts to deduce argument types using local env
// and, when the expression is a call, by consulting known function result types
// (only the first result is considered for scalar param positions).
func inferExprTypeWithEnvAndResults(e ast.Expr, vars map[string]string, results map[string][]string) string {
    switch v := e.(type) {
    case *ast.CallExpr:
        if results != nil {
            if rs, ok := results[v.Name]; ok && len(rs) > 0 && rs[0] != "" { return rs[0] }
        }
        return "any"
    default:
        return inferExprTypeWithVars(e, vars)
    }
}
