package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func inferLocalExprType(env map[string]string, e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        if t := env[v.Name]; t != "" { return t }
        return "any"
    case *ast.NumberLit:
        return "int"
    case *ast.StringLit:
        return "string"
    case *ast.UnaryExpr:
        // logical not yields bool (i1)
        if v.Op == token.Bang { return "bool" }
        return "any"
    case *ast.BinaryExpr:
        xt := inferLocalExprType(env, v.X)
        yt := inferLocalExprType(env, v.Y)
        switch v.Op {
        case token.Plus:
            if xt == "string" && yt == "string" { return "string" }
            if xt == "int" && yt == "int" { return "int" }
        case token.Minus, token.Star, token.Slash, token.Percent:
            if xt == "int" && yt == "int" { return "int" }
        case token.Eq, token.Ne, token.Lt, token.Le, token.Gt, token.Ge:
            // comparisons yield bool regardless of operand type (when comparable)
            return "bool"
        case token.And, token.Or:
            // logical ops yield bool
            return "bool"
        default:
            if xt == yt && xt != "any" { return xt }
        }
        return "any"
    case *ast.SliceLit:
        if v.TypeName != "" { return "slice<" + v.TypeName + ">" }
        if len(v.Elems) == 0 { return "slice<any>" }
        et := inferLocalExprType(env, v.Elems[0])
        if et == "" { et = "any" }
        return "slice<" + et + ">"
    case *ast.SetLit:
        if v.TypeName != "" { return "set<" + v.TypeName + ">" }
        if len(v.Elems) == 0 { return "set<any>" }
        et := inferLocalExprType(env, v.Elems[0])
        if et == "" { et = "any" }
        return "set<" + et + ">"
    case *ast.MapLit:
        kt := v.KeyType
        vt := v.ValType
        if kt == "" { kt = "any" }
        if vt == "" { vt = "any" }
        return "map<" + kt + "," + vt + ">"
    default:
        return "any"
    }
}

