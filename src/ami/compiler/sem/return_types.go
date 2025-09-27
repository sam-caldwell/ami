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
            for _, e := range rs.Results { got = append(got, inferExprType(env, e)) }
            // length mismatch
            if len(got) != len(decl) {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return arity mismatch", Pos: &diag.Position{Line: rs.Pos.Line, Column: rs.Pos.Column, Offset: rs.Pos.Offset}})
                continue
            }
            // element-wise mismatch using compatibility rules
            for i := range got {
                if !typesCompatible(decl[i], got[i]) {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return type mismatch", Pos: &diag.Position{Line: rs.Pos.Line, Column: rs.Pos.Column, Offset: rs.Pos.Offset}})
                    break
                }
            }
        }
    }
    return out
}
