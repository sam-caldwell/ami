package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerBlock lowers a function body block into a sequence of IR instructions.
func lowerBlock(st *lowerState, b *ast.BlockStmt) []ir.Instruction {
    var out []ir.Instruction
    if b == nil { return out }
    for _, s := range b.Stmts {
        switch v := s.(type) {
        case *ast.VarDecl:
            out = append(out, lowerStmtVar(st, v))
        case *ast.AssignStmt:
            out = append(out, lowerStmtAssign(st, v))
        case *ast.ReturnStmt:
            // Materialize return expressions so literals/ops appear as EXPR before RETURN
            var vals []ir.Value
            for _, e := range v.Results {
                if ex, ok := lowerExpr(st, e); ok {
                    if ex.Op != "" || ex.Callee != "" || len(ex.Args) > 0 { out = append(out, ex) }
                    if ex.Result != nil { vals = append(vals, *ex.Result) }
                }
            }
            out = append(out, ir.Return{Values: vals})
        case *ast.DeferStmt:
            out = append(out, lowerStmtDefer(st, v))
        case *ast.ExprStmt:
            // Special-case release(x): emit zeroization call
            if ce, isCall := v.X.(*ast.CallExpr); isCall && ce.Name == "release" && len(ce.Args) >= 1 {
                if exArg, ok := lowerExpr(st, ce.Args[0]); ok {
                    if exArg.Op != "" || exArg.Callee != "" || len(exArg.Args) > 0 { out = append(out, exArg) }
                    // literal 0 length (until size info is available during typing/codegen)
                    zlen := ir.Expr{Op: "lit:0", Result: &ir.Value{ID: st.newTemp(), Type: "int64"}}
                    out = append(out, zlen)
                    // call zeroize(ptr, len)
                    var argv ir.Value
                    if exArg.Result != nil { argv = *exArg.Result } else { argv = ir.Value{ID: "", Type: "ptr"} }
                    out = append(out, ir.Expr{Op: "call", Callee: "ami_rt_zeroize", Args: []ir.Value{argv, {ID: zlen.Result.ID, Type: "int64"}}})
                }
            } else {
                if e, ok := lowerExpr(st, v.X); ok {
                    out = append(out, e)
                }
            }
        }
    }
    return out
}
