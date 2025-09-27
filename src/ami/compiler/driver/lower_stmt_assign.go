package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStmtAssign lowers a simple name = expr assignment into ASSIGN.
func lowerStmtAssign(st *lowerState, as *ast.AssignStmt) ir.Instruction {
    if ex, ok := lowerExpr(st, as.Value); ok && ex.Result != nil {
        // propagate type to dest when not known yet
        if st != nil && st.varTypes != nil && st.varTypes[as.Name] == "" {
            st.varTypes[as.Name] = ex.Result.Type
        }
        return ir.Assign{DestID: as.Name, Src: *ex.Result}
    }
    // fallback: assign from unknown temp value of type any
    id := st.newTemp()
    return ir.Assign{DestID: as.Name, Src: ir.Value{ID: id, Type: "any"}}
}
