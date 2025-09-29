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
                    if ex.Callee == "ami_rt_zeroize" {
                        e.RequireExtern("declare void @ami_rt_zeroize(ptr, i64)")
                    }
                    if ex.Callee == "ami_rt_owned_len" {
                        e.RequireExtern("declare i64 @ami_rt_owned_len(ptr)")
                    }
                    if ex.Callee == "ami_rt_owned_new" {
                        e.RequireExtern("declare ptr @ami_rt_owned_new(i8*, i64)")
                    }
                    if ex.Callee == "ami_rt_zeroize_owned" {
                        e.RequireExtern("declare void @ami_rt_zeroize_owned(ptr)")
                    }
                } else if d, ok := ins.(ir.Defer); ok {
                    ex := d.Expr
                    if strings.ToLower(ex.Op) == "call" {
                        switch ex.Callee {
                        case "ami_rt_panic":
                            e.RequireExtern("declare void @ami_rt_panic(i32)")
                        case "ami_rt_alloc":
                            e.RequireExtern("declare ptr @ami_rt_alloc(i64)")
                        case "ami_rt_zeroize":
                            e.RequireExtern("declare void @ami_rt_zeroize(ptr, i64)")
                        case "ami_rt_owned_len":
                            e.RequireExtern("declare i64 @ami_rt_owned_len(ptr)")
                        case "ami_rt_owned_new":
                            e.RequireExtern("declare ptr @ami_rt_owned_new(i8*, i64)")
                        case "ami_rt_zeroize_owned":
                            e.RequireExtern("declare void @ami_rt_zeroize_owned(ptr)")
                        }
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

// EmitModuleLLVMForTarget lowers an IR module to LLVM IR using a specific target triple.
func EmitModuleLLVMForTarget(m ir.Module, triple string) (string, error) {
    e := NewModuleEmitter(m.Package, "")
    if triple != "" { e.SetTargetTriple(triple) }
    // Collect externs based on usage (same as EmitModuleLLVM)
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
                    if ex.Callee == "ami_rt_zeroize" {
                        e.RequireExtern("declare void @ami_rt_zeroize(ptr, i64)")
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
