package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// isStringLit reports whether e is a string literal.
func isStringLit(e ast.Expr) bool {
    _, ok := e.(*ast.StringLit)
    return ok
}

