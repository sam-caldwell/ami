package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStmtDefer lowers a defer call into a DEFER instruction carrying an Expr.
func lowerStmtDefer(st *lowerState, ds *ast.DeferStmt) ir.Instruction {
    // If this is defer release(x), emit a zeroize-owned call to preserve semantics at exit.
    if ds != nil && ds.Call != nil && ds.Call.Name == "release" && len(ds.Call.Args) >= 1 {
        // Lower the argument to obtain a value id
        if exArg, ok := lowerExpr(st, ds.Call.Args[0]); ok {
            var argv ir.Value
            if exArg.Result != nil { argv = *exArg.Result } else { argv = ir.Value{ID: "", Type: "ptr"} }
            // Build a call expression: ami_rt_zeroize_owned(x)
            z := ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{argv}}
            return ir.Defer{Expr: z}
        }
    }
    // lower call expression to ir.Expr without capturing the result
    ex := lowerCallExpr(st, ds.Call)
    ex.Result = nil
    return ir.Defer{Expr: ex}
}
