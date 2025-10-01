package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeReturnTypes compares declared function result types with return statements.
// Emits E_RETURN_TYPE_MISMATCH on length/type mismatches (scaffold typing rules).
func AnalyzeReturnTypes(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok { continue }
        decl := make([]string, len(fn.Results))
        for i, r := range fn.Results { decl[i] = r.Type }
        // scan body for return statements
        if fn.Body == nil { continue }
        env := buildLocalEnv(fn)
        for _, st := range fn.Body.Stmts {
            rs, ok := st.(*ast.ReturnStmt)
            if !ok { continue }
            // infer result types
            var got []string
            var gpos []diag.Position
            for _, e := range rs.Results {
                got = append(got, inferExprType(env, e))
                p := epos(e)
                gpos = append(gpos, diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset})
            }
            // length mismatch
            if len(got) != len(decl) {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return arity mismatch", Pos: &diag.Position{Line: rs.Pos.Line, Column: rs.Pos.Column, Offset: rs.Pos.Offset}})
                continue
            }
            // element-wise mismatch using compatibility rules
            var mismatchIndices []int
            for i := range got {
                // compute positions for precision
                expPos := diag.Position{}
                if i < len(fn.Results) { expPos = diag.Position{Line: fn.Results[i].Pos.Line, Column: fn.Results[i].Pos.Column, Offset: fn.Results[i].Pos.Offset} }
                actPos := gpos[i]
                if mismatch, path, pathIdx, fieldPath, base, wantN, gotN := findGenericArityMismatchWithFields(decl[i], got[i]); mismatch {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_GENERIC_ARITY_MISMATCH", Message: "generic type argument count mismatch", Pos: &actPos, Data: map[string]any{"index": i, "base": base, "path": path, "pathIdx": pathIdx, "fieldPath": fieldPath, "expected": decl[i], "actual": got[i], "expectedArity": wantN, "actualArity": gotN, "expectedPos": expPos}})
                    mismatchIndices = append(mismatchIndices, i)
                    // continue checking remaining positions to surface multiple issues
                    continue
                } else if mismatch, path, pathIdx, base, wantN, gotN := findGenericArityMismatchDeepPath(decl[i], got[i]); mismatch {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_GENERIC_ARITY_MISMATCH", Message: "generic type argument count mismatch", Pos: &actPos, Data: map[string]any{"index": i, "base": base, "path": path, "pathIdx": pathIdx, "expected": decl[i], "actual": got[i], "expectedArity": wantN, "actualArity": gotN, "expectedPos": expPos}})
                    mismatchIndices = append(mismatchIndices, i)
                    continue
                }
                if !typesCompatible(decl[i], got[i]) {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return type mismatch", Pos: &actPos, Data: map[string]any{"index": i, "expected": decl[i], "actual": got[i], "expectedPos": expPos}})
                    mismatchIndices = append(mismatchIndices, i)
                }
            }
            if len(mismatchIndices) > 1 {
                var paths []map[string]any
                for _, i := range mismatchIndices {
                    if i < len(decl) && i < len(got) {
                        if m, p, idx, fp, b, _, _ := findGenericArityMismatchWithFields(decl[i], got[i]); m {
                            paths = append(paths, map[string]any{"index": i, "base": b, "path": p, "pathIdx": idx, "fieldPath": fp})
                        }
                    }
                }
                data := map[string]any{"count": len(mismatchIndices), "indices": mismatchIndices}
                if len(paths) > 0 { data["paths"] = paths }
                // emit a summary aggregate alongside per-element diagnostics
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RETURN_TUPLE_MISMATCH_SUMMARY", Message: "multiple return elements mismatch", Pos: &diag.Position{Line: rs.Pos.Line, Column: rs.Pos.Column, Offset: rs.Pos.Offset}, Data: data})
            }
        }
    }
    return out
}
