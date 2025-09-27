package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStmtVar lowers a variable declaration to a VAR instruction with optional init.
func lowerStmtVar(st *lowerState, vd *ast.VarDecl) ir.Instruction {
    var init *ir.Value
    if vd.Init != nil {
        if ex, ok := lowerExpr(st, vd.Init); ok && ex.Result != nil {
            // materialize a value for init
            v := *ex.Result
            init = &v
        }
    }
    // create a symbolic variable value id equal to the variable name
    res := ir.Value{ID: vd.Name, Type: vd.Type}
    if st != nil && st.varTypes != nil && vd.Name != "" && vd.Type != "" {
        st.varTypes[vd.Name] = vd.Type
    }
    return ir.Var{Name: vd.Name, Type: vd.Type, Init: init, Result: res}
}
