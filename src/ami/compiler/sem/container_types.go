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

func prim(t string) bool {
    switch t {
    case "", "any":
        return false
    case "bool", "byte", "rune",
        "int", "int8", "int16", "int32", "int64", "int128",
        "uint", "uint8", "uint16", "uint32", "uint64", "uint128",
        "float32", "float64",
        "string":
        return true
    default:
        return false
    }
}

func deduceType(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.NumberLit:
        return "int"
    case *ast.StringLit:
        return "string"
    case *ast.SliceLit:
        if v.TypeName != "" { return "slice<" + v.TypeName + ">" }
        return "slice<any>"
    case *ast.SetLit:
        if v.TypeName != "" { return "set<" + v.TypeName + ">" }
        return "set<any>"
    case *ast.MapLit:
        kt := v.KeyType; vt := v.ValType
        if kt == "" { kt = "any" }
        if vt == "" { vt = "any" }
        return "map<" + kt + "," + vt + ">"
    default:
        return "any"
    }
}
