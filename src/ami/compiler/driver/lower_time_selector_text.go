package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func selectorText(s *ast.SelectorExpr) string {
    if s == nil { return "" }
    left := ""
    switch v := s.X.(type) {
    case *ast.IdentExpr:
        left = v.Name
    case *ast.SelectorExpr:
        left = selectorText(v)
    default:
        left = "?"
    }
    if left == "" { left = "?" }
    return left + "." + s.Sel
}

