package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStmtVar lowers a variable declaration to a VAR instruction with optional init.
func lowerStmtVar(st *lowerState, vd *ast.VarDecl) ir.Instruction {
    var init *ir.Value
    var vtype string = vd.Type
    if vd.Init != nil {
        if ex, ok := lowerExpr(st, vd.Init); ok && ex.Result != nil {
            // materialize a value for init
            v := *ex.Result
            init = &v
            // if no explicit type, infer from initializer
            if vtype == "" && v.Type != "" && v.Type != "any" {
                vtype = v.Type
            }
        }
    }
    // create a symbolic variable value id equal to the variable name
    res := ir.Value{ID: vd.Name, Type: vtype}
    if st != nil && st.varTypes != nil && vd.Name != "" && vtype != "" {
        st.varTypes[vd.Name] = vtype
    }
    return ir.Var{Name: vd.Name, Type: vtype, Init: init, Result: res}
}
