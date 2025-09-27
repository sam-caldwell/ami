package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeReturnInference infers return types for functions without declared
// result types by inspecting return statements. It emits:
// - E_RETURN_TYPE_MISMATCH on inconsistent arity/types across returns
// - E_TYPE_UNINFERRED when any result type remains unknown (any)
func AnalyzeReturnInference(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        if len(fn.Results) > 0 { continue }
        // Build env for inference
        env := buildLocalEnv(fn)
        var proto []string
        consistent := true
        uninferred := false
        // Gather concrete from returns
        for _, st := range fn.Body.Stmts {
            rs, ok := st.(*ast.ReturnStmt)
            if !ok { continue }
            // infer types for this return
            var got []string
            for _, e := range rs.Results { got = append(got, inferRetTypeWithEnv(e, env, collectFunctionResults(f))) }
            if len(proto) == 0 {
                proto = got
                continue
            }
            if len(got) != len(proto) { consistent = false; break }
            for i := range got {
                if !typesCompatible(proto[i], got[i]) {
                    consistent = false
                    break
                }
            }
            if !consistent { break }
        }
        if !consistent {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "inconsistent inferred return types", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
            continue
        }
        // Check for uninferred (any) results
        for _, t := range proto { if t == "" || t == "any" { uninferred = true } }
        if uninferred {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_UNINFERRED", Message: "unable to infer concrete return type(s)", Pos: &diag.Position{Line: fn.NamePos.Line, Column: fn.NamePos.Column, Offset: fn.NamePos.Offset}})
        }
    }
    return out
}

