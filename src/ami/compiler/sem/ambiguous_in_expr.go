package sem

import (
    "time"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

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

