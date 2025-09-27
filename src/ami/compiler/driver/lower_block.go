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
            if e, ok := lowerExpr(st, v.X); ok {
                out = append(out, e)
            }
        }
    }
    return out
}
