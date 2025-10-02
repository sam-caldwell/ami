package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func exprOffset(e ast.Expr) int {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Pos.Offset
    case *ast.CallExpr:
        return v.Pos.Offset
    case *ast.SelectorExpr:
        return v.Pos.Offset
    case *ast.StringLit:
        return v.Pos.Offset
    case *ast.NumberLit:
        return v.Pos.Offset
    default:
        return -1
    }
}

