package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeAmbiguity reports E_TYPE_AMBIGUOUS for container literals that cannot be typed.
func AnalyzeAmbiguity(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        for _, s := range fn.Body.Stmts {
            switch v := s.(type) {
            case *ast.VarDecl:
                if v.Init != nil { out = append(out, ambiguousInExpr(now, v.Init)...) }
            case *ast.AssignStmt:
                if v.Value != nil { out = append(out, ambiguousInExpr(now, v.Value)...) }
            case *ast.ReturnStmt:
                for _, e := range v.Results { out = append(out, ambiguousInExpr(now, e)...) }
            case *ast.ExprStmt:
                if v.X != nil { out = append(out, ambiguousInExpr(now, v.X)...) }
            }
        }
    }
    return out
}
