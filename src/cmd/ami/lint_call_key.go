package main

import ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func callKey(prefix string, ce *ast.CallExpr) string {
    if ce == nil { return "" }
    return prefix + "@" + ce.Name + ":" + itoa(ce.NamePos.Line) + ":" + itoa(ce.NamePos.Column)
}

