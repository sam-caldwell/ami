package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

func checkContainerExpr(now time.Time, e ast.Expr) []diag.Record {
    var out []diag.Record
    switch v := e.(type) {
    case *ast.SliceLit:
        base := ""
        for _, el := range v.Elems {
            t := deduceType(el)
            if base == "" && t != "any" { base = t }
            if t != "any" && base != "" && t != base {
                p := epos(el)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "container element type mismatch: expected " + base + ", got " + t, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
            }
        }
        // If declared primitive type is present, ensure it matches the base when known.
        if prim(v.TypeName) && base != "" && base != v.TypeName {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "slice element type does not match declared type " + v.TypeName, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
        }
    case *ast.SetLit:
        base := ""
        for _, el := range v.Elems {
            t := deduceType(el)
            if base == "" && t != "any" { base = t }
            if t != "any" && base != "" && t != base {
                p := epos(el)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "set element type mismatch: expected " + base + ", got " + t, Pos: &diag.Position{Line: p.Line, Column: p.Column, Offset: p.Offset}})
            }
        }
        if prim(v.TypeName) && base != "" && base != v.TypeName {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "set element type does not match declared type " + v.TypeName, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
        }
    case *ast.MapLit:
        kbase := ""
        vbase := ""
        for _, kv := range v.Elems {
            kt := deduceType(kv.Key)
            vt := deduceType(kv.Val)
            if kbase == "" && kt != "any" { kbase = kt }
            if vbase == "" && vt != "any" { vbase = vt }
            if kt != "any" && kbase != "" && kt != kbase {
                pos := epos(kv.Key)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map key type mismatch: expected " + kbase + ", got " + kt, Pos: &diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset}})
            }
            if vt != "any" && vbase != "" && vt != vbase {
                pos := epos(kv.Val)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map value type mismatch: expected " + vbase + ", got " + vt, Pos: &diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset}})
            }
        }
        if prim(v.KeyType) && kbase != "" && kbase != v.KeyType {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map key type does not match declared type " + v.KeyType, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
        }
        if prim(v.ValType) && vbase != "" && vbase != v.ValType {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map value type does not match declared type " + v.ValType, Pos: &diag.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
        }
    }
    return out
}

