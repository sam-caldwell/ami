package driver

import (
    "hash/fnv"
    "strconv"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// handlerTokenValue computes the deterministic token value for a handler expression.
func handlerTokenValue(e ast.Expr) (int64, bool) {
    name := ""
    switch v := e.(type) {
    case *ast.IdentExpr:
        name = v.Name
    case *ast.SelectorExpr:
        name = selectorText(v)
    }
    if name != "" {
        h := fnv.New64a(); _, _ = h.Write([]byte(name))
        return int64(h.Sum64()), true
    }
    off := exprOffset(e)
    if off >= 0 {
        h := fnv.New64a(); _, _ = h.Write([]byte("anon@" + strconv.Itoa(off)))
        return int64(h.Sum64()), true
    }
    return 0, false
}

