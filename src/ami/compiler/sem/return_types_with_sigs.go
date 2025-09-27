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
        for _, st := range fn.Body.Stmts {
            rs, ok := st.(*ast.ReturnStmt)
            if !ok { continue }
            // infer result types
            var got []string
            for _, e := range rs.Results { got = append(got, inferRetTypeWithSigs(e, results)) }
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

func inferRetTypeWithSigs(e ast.Expr, results map[string][]string) string {
    switch v := e.(type) {
    case *ast.CallExpr:
        if rs, ok := results[v.Name]; ok && len(rs) > 0 && rs[0] != "" { return rs[0] }
        return "any"
    default:
        return inferExprType(e)
    }
}

