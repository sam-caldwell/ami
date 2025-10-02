package llvm

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// moved to emit_module.go

// EmitModuleLLVMForTarget lowers an IR module to LLVM IR using a specific target triple.
func EmitModuleLLVMForTarget(m ir.Module, triple string) (string, error) {
    e := NewModuleEmitter(m.Package, "")
    if triple != "" { e.SetTargetTriple(triple) }
    // Collect externs based on usage
    for _, f := range m.Functions {
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                switch v := ins.(type) {
                case ir.Expr:
                    addExternsForExpr(e, v)
                case ir.Defer:
                    addExternsForExpr(e, v.Expr)
                }
            }
        }
    }
    for _, f := range m.Functions {
        if err := e.AddFunction(f); err != nil { return "", err }
    }
    return e.Build(), nil
}

// addExternsForExpr moved to add_externs_for_expr.go to satisfy single-declaration rule
