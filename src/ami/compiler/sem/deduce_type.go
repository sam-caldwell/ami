package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func deduceType(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.NumberLit:
        return "int"
    case *ast.StringLit:
        return "string"
    case *ast.BinaryExpr:
        // simple arithmetic infer: if both sides are int and op is arithmetic, infer int
        x := deduceType(v.X)
        y := deduceType(v.Y)
        switch v.Op {
        case token.Plus, token.Minus, token.Star, token.Slash, token.Percent:
            if x == "int" && y == "int" { return "int" }
        }
        return "any"
    case *ast.SliceLit:
        if v.TypeName != "" { return "slice<" + v.TypeName + ">" }
        // infer from first element when available
        if len(v.Elems) > 0 {
            et := deduceType(v.Elems[0])
            if et == "" { et = "any" }
            return "slice<" + et + ">"
        }
        return "slice<any>"
    case *ast.SetLit:
        if v.TypeName != "" { return "set<" + v.TypeName + ">" }
        if len(v.Elems) > 0 {
            et := deduceType(v.Elems[0])
            if et == "" { et = "any" }
            return "set<" + et + ">"
        }
        return "set<any>"
    case *ast.MapLit:
        kt := v.KeyType; vt := v.ValType
        if len(v.Elems) > 0 {
            // infer from first pair
            kt0 := deduceType(v.Elems[0].Key)
            vt0 := deduceType(v.Elems[0].Val)
            if kt == "" { kt = kt0 }
            if vt == "" { vt = vt0 }
        }
        if kt == "" { kt = "any" }
        if vt == "" { vt = "any" }
        return "map<" + kt + "," + vt + ">"
    default:
        return "any"
    }
}

