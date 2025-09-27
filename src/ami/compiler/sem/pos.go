package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// epos returns a best-effort source position for an expression.
func epos(e ast.Expr) source.Position {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Pos
    case *ast.StringLit:
        return v.Pos
    case *ast.NumberLit:
        return v.Pos
    case *ast.CallExpr:
        return v.Pos
    case *ast.BinaryExpr:
        return v.Pos
    case *ast.SliceLit:
        return v.Pos
    case *ast.SetLit:
        return v.Pos
    case *ast.MapLit:
        return v.Pos
    case *ast.SelectorExpr:
        return v.Pos
    default:
        return source.Position{}
    }
}

