package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStmtAssign lowers a simple name = expr assignment into ASSIGN.
func lowerStmtAssign(st *lowerState, as *ast.AssignStmt) ir.Instruction {
    if ex, ok := lowerExpr(st, as.Value); ok && ex.Result != nil {
        return ir.Assign{DestID: as.Name, Src: *ex.Result}
    }
    // fallback: assign from unknown value of type any
    return ir.Assign{DestID: as.Name, Src: ir.Value{ID: "", Type: "any"}}
}

