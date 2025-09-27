package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeAmbiguity reports E_TYPE_AMBIGUOUS for container literals that cannot
// be typed due to missing type arguments and empty elements (or effectively any).
// Emits precise positions pointing to the literal site.
func AnalyzeAmbiguity(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // walk function bodies
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

func ambiguousInExpr(now time.Time, e ast.Expr) []diag.Record {
    var out []diag.Record
    switch v := e.(type) {
    case *ast.SliceLit:
        if v.TypeName == "" || v.TypeName == "any" {
            if len(v.Elems) == 0 {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_AMBIGUOUS", Message: "ambiguous slice literal: no type and no elements", Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
            }
        }
    case *ast.SetLit:
        if v.TypeName == "" || v.TypeName == "any" {
            if len(v.Elems) == 0 {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_AMBIGUOUS", Message: "ambiguous set literal: no type and no elements", Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
            }
        }
    case *ast.MapLit:
        if (v.KeyType == "" || v.KeyType == "any") && (v.ValType == "" || v.ValType == "any") {
            if len(v.Elems) == 0 {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_AMBIGUOUS", Message: "ambiguous map literal: no type and no elements", Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
            }
        }
    case *ast.BinaryExpr:
        out = append(out, ambiguousInExpr(now, v.X)...)
        out = append(out, ambiguousInExpr(now, v.Y)...)
    case *ast.CallExpr:
        for _, a := range v.Args { out = append(out, ambiguousInExpr(now, a)...) }
    }
    return out
}
