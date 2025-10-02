package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func flattenSelector(s *ast.SelectorExpr) (base string, path string) {
    parts := []string{}
    var cur ast.Expr = s
    for {
        es, ok := cur.(*ast.SelectorExpr)
        if !ok { break }
        parts = append([]string{es.Sel}, parts...)
        cur = es.X
    }
    if id, ok := cur.(*ast.IdentExpr); ok {
        base = id.Name
    }
    if len(parts) > 0 { path = strings.Join(parts, ".") }
    return base, path
}

