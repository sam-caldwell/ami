package llvm

import (
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// EmitModuleLLVM lowers an IR module to textual LLVM IR string.
// It supports a minimal subset sufficient for initial golden tests.
func EmitModuleLLVM(m ir.Module) (string, error) {
    e := NewModuleEmitter(m.Package, "")
    // Collect externs based on usage (scaffold): panic/alloc
    for _, f := range m.Functions {
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                if ex, ok := ins.(ir.Expr); ok {
                    op := strings.ToLower(ex.Op)
                    if op == "panic" {
                        e.RequireExtern("declare void @ami_rt_panic(i32)")
                    }
                    if op == "alloc" || ex.Callee == "ami_rt_alloc" {
                        e.RequireExtern("declare ptr @ami_rt_alloc(i64)")
                    }
                }
            }
        }
    }
    for _, f := range m.Functions {
        if err := e.AddFunction(f); err != nil { return "", err }
    }
    return e.Build(), nil
}
