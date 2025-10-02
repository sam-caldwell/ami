package driver

import (
    "hash/fnv"
    "strconv"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// handlerTokenImmediate returns an immediate ID ("#<num>") for a handler expression.
func handlerTokenImmediate(e ast.Expr) (string, bool) {
    name := ""
    switch v := e.(type) {
    case *ast.IdentExpr:
        name = v.Name
    case *ast.SelectorExpr:
        name = selectorText(v)
    }
    if name != "" {
        h := fnv.New64a(); _, _ = h.Write([]byte(name))
        tok := int64(h.Sum64())
        return "#" + strconv.FormatInt(tok, 10), true
    }
    off := exprOffset(e)
    if off >= 0 {
        h := fnv.New64a(); _, _ = h.Write([]byte("anon@" + strconv.Itoa(off)))
        tok := int64(h.Sum64())
        return "#" + strconv.FormatInt(tok, 10), true
    }
    return "", false
}

