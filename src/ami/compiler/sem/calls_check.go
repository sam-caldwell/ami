package sem

import (
    "fmt"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

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
        } else if mismatch, path, pathIdx, fieldPath, base, wantN, gotN := findGenericArityMismatchDeepPathTextFields(pt, at); mismatch {
            p := epos(a)
            data := map[string]any{"argIndex": i, "callee": c.Name, "base": base, "path": path, "pathIdx": pathIdx, "expected": pt, "actual": at, "expectedArity": wantN, "actualArity": gotN}
            if len(fieldPath) > 0 { data["fieldPath"] = fieldPath }
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
                e := map[string]any{"argIndex": i, "base": b, "path": p, "pathIdx": idx, "fieldPath": fp}
                if i < len(s.paramTypePos) { e["expectedPos"] = s.paramTypePos[i] }
                if i < len(s.paramNames) && s.paramNames[i] != "" { e["paramName"] = s.paramNames[i] }
                paths = append(paths, e)
            } else if m2, p2, idx2, fp2, b2, _, _ := findGenericArityMismatchDeepPathTextFields(pt, at); m2 {
                e := map[string]any{"argIndex": i, "base": b2, "path": p2, "pathIdx": idx2}
                if len(fp2) > 0 { e["fieldPath"] = fp2 }
                if i < len(s.paramTypePos) { e["expectedPos"] = s.paramTypePos[i] }
                if i < len(s.paramNames) && s.paramNames[i] != "" { e["paramName"] = s.paramNames[i] }
                paths = append(paths, e)
            }
        }
        data := map[string]any{"count": len(mismatchIdx), "indices": mismatchIdx, "callee": c.Name}
        if len(paths) > 0 { data["paths"] = paths }
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CALL_ARGS_MISMATCH_SUMMARY", Message: "multiple call arguments mismatch", Pos: &diag.Position{Line: c.Pos.Line, Column: c.Pos.Column, Offset: c.Pos.Offset}, Data: data})
    }
    return out
}

