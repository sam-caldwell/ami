package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeContainerTypes enforces consistent element/key/value types within
// container literals and performs simple checks against declared primitive
// type parameters (when present). It emits E_TYPE_MISMATCH with positions.
func AnalyzeContainerTypes(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // Walk function bodies to find literals in var inits and assignments.
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        for _, s := range fn.Body.Stmts {
            switch v := s.(type) {
            case *ast.VarDecl:
                if v.Init == nil { continue }
                out = append(out, checkContainerExpr(now, v.Init)...)
            case *ast.AssignStmt:
                if v.Value == nil { continue }
                out = append(out, checkContainerExpr(now, v.Value)...)
            }
        }
    }
    return out
}


// containerCompatibleWith compares a declared container type string against a literal expression's
// inferred element/key/value base types.
// containerCompatibleWith: removed (parser does not capture generics for result types currently).
