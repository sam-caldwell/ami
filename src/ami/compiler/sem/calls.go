package sem

import (
    "fmt"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

type sig struct{ params, results []string; paramNames []string; paramTypePos []diag.Position }

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
    var mismatchIdx []int
    for i, a := range c.Args {
        at := inferExprTypeWithVars(a, vars)
        pt := s.params[i]
        if pt == "" || pt == "any" || at == "any" { continue }
        // Prefer a specific generic arity mismatch diagnostic when applicable
        // Try typed (richer) detection first; fall back to textual deep detection
        if mismatch, path, pathIdx, fieldPath, base, wantN, gotN := findGenericArityMismatchWithFields(pt, at); mismatch {
            p := epos(a)
            data := map[string]any{"argIndex": i, "callee": c.Name, "base": base, "path": path, "pathIdx": pathIdx, "fieldPath": fieldPath, "expected": pt, "actual": at, "expectedArity": wantN, "actualArity": gotN}
            if i < len(s.paramNames) && s.paramNames[i] != "" { data["paramName"] = s.paramNames[i] }
            if i < len(s.paramTypePos) { data["expectedPos"] = s.paramTypePos[i] }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_GENERIC_ARITY_MISMATCH", Message: "generic type argument count mismatch", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: data})
            mismatchIdx = append(mismatchIdx, i)
            continue
        } else if mismatch, path, pathIdx, base, wantN, gotN := findGenericArityMismatchDeepPath(pt, at); mismatch {
            p := epos(a)
            data := map[string]any{"argIndex": i, "callee": c.Name, "base": base, "path": path, "pathIdx": pathIdx, "expected": pt, "actual": at, "expectedArity": wantN, "actualArity": gotN}
            if i < len(s.paramNames) && s.paramNames[i] != "" { data["paramName"] = s.paramNames[i] }
            if i < len(s.paramTypePos) { data["expectedPos"] = s.paramTypePos[i] }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_GENERIC_ARITY_MISMATCH", Message: "generic type argument count mismatch", Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: data})
            mismatchIdx = append(mismatchIdx, i)
            continue
        }
        if !typesCompatible(pt, at) {
            p := epos(a)
            msg := fmt.Sprintf("call argument type mismatch: arg %d expected %s, got %s", i, pt, at)
            data := map[string]any{"argIndex": i, "expected": pt, "actual": at, "callee": c.Name}
            if i < len(s.paramNames) && s.paramNames[i] != "" { data["paramName"] = s.paramNames[i] }
            if i < len(s.paramTypePos) { data["expectedPos"] = s.paramTypePos[i] }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARG_TYPE_MISMATCH", Message: msg, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}, Data: data})
            mismatchIdx = append(mismatchIdx, i)
        }
    }
    if len(mismatchIdx) > 1 {
        // include shallow path hints by recomputing for mismatching args
        var paths []map[string]any
        for _, i := range mismatchIdx {
            pt := s.params[i]
            at := inferExprTypeWithVars(c.Args[i], vars)
            if m, p, idx, fp, b, _, _ := findGenericArityMismatchWithFields(pt, at); m {
                paths = append(paths, map[string]any{"argIndex": i, "base": b, "path": p, "pathIdx": idx, "fieldPath": fp})
            } else if m2, p2, idx2, b2, _, _ := findGenericArityMismatchDeepPath(pt, at); m2 {
                paths = append(paths, map[string]any{"argIndex": i, "base": b2, "path": p2, "pathIdx": idx2})
            }
        }
        data := map[string]any{"count": len(mismatchIdx), "indices": mismatchIdx, "callee": c.Name}
        if len(paths) > 0 { data["paths"] = paths }
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARGS_MISMATCH_SUMMARY", Message: "multiple call arguments mismatch", Pos: &diag.Position{Line: c.Pos.Line, Column: c.Pos.Column, Offset: c.Pos.Offset}, Data: data})
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
