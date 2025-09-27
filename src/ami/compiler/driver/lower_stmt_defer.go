package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStmtDefer lowers a defer call into a DEFER instruction carrying an Expr.
func lowerStmtDefer(st *lowerState, ds *ast.DeferStmt) ir.Instruction {
    // lower call expression to ir.Expr without capturing the result
    ex := lowerCallExpr(st, ds.Call)
    ex.Result = nil
    return ir.Defer{Expr: ex}
}

