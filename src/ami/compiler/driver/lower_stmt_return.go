package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStmtReturn lowers a return statement into RETURN with values.
func lowerStmtReturn(st *lowerState, rs *ast.ReturnStmt) ir.Instruction {
    var vals []ir.Value
    for _, e := range rs.Results {
        if ex, ok := lowerExpr(st, e); ok && ex.Result != nil {
            vals = append(vals, *ex.Result)
        }
    }
    return ir.Return{Values: vals}
}

