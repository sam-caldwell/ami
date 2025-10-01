package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// exprText produces a simple string representation of an expression suitable
// for debug displays (attribute args, etc.). It is not a full pretty-printer.
func exprText(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Name
    case *ast.StringLit:
        return v.Value
    case *ast.DurationLit:
        return v.Text
    case *ast.NumberLit:
        return v.Text
    case *ast.SelectorExpr:
        left := exprText(v.X)
        if left == "" {
            left = "?"
        }
        return left + "." + v.Sel
    case *ast.CallExpr:
        // return callee name with parentheses; include ellipsis when args present
        if len(v.Args) > 0 {
            return v.Name + "(…)"
        }
        return v.Name + "()"
    case *ast.SliceLit:
        return "slice"
    case *ast.SetLit:
        return "set"
    case *ast.MapLit:
        return "map"
    default:
        return ""
    }
}

