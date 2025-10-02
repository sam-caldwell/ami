package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// emitNestedCallArgs lowers nested call expressions used as arguments before the outer call.
func emitNestedCallArgs(st *lowerState, e ast.Expr, out *[]ir.Instruction) {
    var walkArgs func(ast.Expr)
    walkArgs = func(x ast.Expr) {
        if ce, ok := x.(*ast.CallExpr); ok {
            for _, a := range ce.Args { walkArgs(a) }
            if ex, ok2 := lowerExpr(st, ce); ok2 {
                if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 || len(ex.Results) > 0 { *out = append(*out, ex) }
            }
            return
        }
    }
    if ce, ok := e.(*ast.CallExpr); ok {
        for _, a := range ce.Args { walkArgs(a) }
    }
}

