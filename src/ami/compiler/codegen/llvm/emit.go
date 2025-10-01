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
                if ex.Callee == "ami_rt_string_len" {
                    e.RequireExtern("declare i64 @ami_rt_string_len(ptr)")
                }
                if ex.Callee == "ami_rt_slice_len" {
                    e.RequireExtern("declare i64 @ami_rt_slice_len(ptr)")
                }
                if ex.Callee == "ami_rt_owned_ptr" {
                    e.RequireExtern("declare ptr @ami_rt_owned_ptr(ptr)")
                }
                if ex.Callee == "ami_rt_owned_new" {
                    e.RequireExtern("declare ptr @ami_rt_owned_new(i8*, i64)")
                }
                if ex.Callee == "ami_rt_zeroize_owned" {
                    e.RequireExtern("declare void @ami_rt_zeroize_owned(ptr)")
                }
                if ex.Callee == "ami_rt_sleep_ms" {
                    e.RequireExtern("declare void @ami_rt_sleep_ms(i64)")
                }
                if ex.Callee == "ami_rt_time_now" {
                    e.RequireExtern("declare i64 @ami_rt_time_now()")
                }
                if ex.Callee == "ami_rt_time_add" {
                    e.RequireExtern("declare i64 @ami_rt_time_add(i64, i64)")
                }
                if ex.Callee == "ami_rt_time_delta" {
                    e.RequireExtern("declare i64 @ami_rt_time_delta(i64, i64)")
                }
                if ex.Callee == "ami_rt_time_unix" {
                    e.RequireExtern("declare i64 @ami_rt_time_unix(i64)")
                }
                if ex.Callee == "ami_rt_time_unix_nano" {
                    e.RequireExtern("declare i64 @ami_rt_time_unix_nano(i64)")
                }
                if ex.Callee == "ami_rt_signal_register" {
                    // Handler is represented as an opaque i64 token (no raw ptr exposure)
                    e.RequireExtern("declare void @ami_rt_signal_register(i64, i64)")
                }
                if ex.Callee == "ami_rt_os_signal_enable" {
                    e.RequireExtern("declare void @ami_rt_os_signal_enable(i64)")
                }
                if ex.Callee == "ami_rt_os_signal_disable" {
                    e.RequireExtern("declare void @ami_rt_os_signal_disable(i64)")
                }
                if ex.Callee == "ami_rt_install_handler_thunk" {
                    e.RequireExtern("declare void @ami_rt_install_handler_thunk(i64, ptr)")
                }
                if ex.Callee == "ami_rt_get_handler_thunk" {
                    e.RequireExtern("declare ptr @ami_rt_get_handler_thunk(i64)")
                }
                if ex.Callee == "ami_rt_posix_install_trampoline" {
                    e.RequireExtern("declare void @ami_rt_posix_install_trampoline(i64)")
                }
                if ex.Callee == "ami_rt_gpu_blocking_submit" {
                    e.RequireExtern("declare ptr @ami_rt_gpu_blocking_submit(ptr)")
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
                        case "ami_rt_string_len":
                            e.RequireExtern("declare i64 @ami_rt_string_len(ptr)")
                        case "ami_rt_slice_len":
                            e.RequireExtern("declare i64 @ami_rt_slice_len(ptr)")
                        case "ami_rt_owned_ptr":
                            e.RequireExtern("declare ptr @ami_rt_owned_ptr(ptr)")
                        case "ami_rt_owned_new":
                            e.RequireExtern("declare ptr @ami_rt_owned_new(i8*, i64)")
                        case "ami_rt_zeroize_owned":
                            e.RequireExtern("declare void @ami_rt_zeroize_owned(ptr)")
                        case "ami_rt_sleep_ms":
                            e.RequireExtern("declare void @ami_rt_sleep_ms(i64)")
                        case "ami_rt_time_now":
                            e.RequireExtern("declare i64 @ami_rt_time_now()")
                        case "ami_rt_time_add":
                            e.RequireExtern("declare i64 @ami_rt_time_add(i64, i64)")
                        case "ami_rt_time_delta":
                            e.RequireExtern("declare i64 @ami_rt_time_delta(i64, i64)")
                        case "ami_rt_time_unix":
                            e.RequireExtern("declare i64 @ami_rt_time_unix(i64)")
                        case "ami_rt_time_unix_nano":
                            e.RequireExtern("declare i64 @ami_rt_time_unix_nano(i64)")
                        case "ami_rt_signal_register":
                            e.RequireExtern("declare void @ami_rt_signal_register(i64, i64)")
                        case "ami_rt_os_signal_enable":
                            e.RequireExtern("declare void @ami_rt_os_signal_enable(i64)")
                        case "ami_rt_os_signal_disable":
                            e.RequireExtern("declare void @ami_rt_os_signal_disable(i64)")
                        case "ami_rt_install_handler_thunk":
                            e.RequireExtern("declare void @ami_rt_install_handler_thunk(i64, ptr)")
                        case "ami_rt_get_handler_thunk":
                            e.RequireExtern("declare ptr @ami_rt_get_handler_thunk(i64)")
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
                    if op == "panic" { e.RequireExtern("declare void @ami_rt_panic(i32)") }
                    if op == "alloc" || ex.Callee == "ami_rt_alloc" { e.RequireExtern("declare ptr @ami_rt_alloc(i64)") }
                    switch ex.Callee {
                    case "ami_rt_zeroize":
                        e.RequireExtern("declare void @ami_rt_zeroize(ptr, i64)")
                    case "ami_rt_owned_len":
                        e.RequireExtern("declare i64 @ami_rt_owned_len(ptr)")
                    case "ami_rt_owned_ptr":
                        e.RequireExtern("declare ptr @ami_rt_owned_ptr(ptr)")
                    case "ami_rt_owned_new":
                        e.RequireExtern("declare ptr @ami_rt_owned_new(i8*, i64)")
                    case "ami_rt_zeroize_owned":
                        e.RequireExtern("declare void @ami_rt_zeroize_owned(ptr)")
                    case "ami_rt_sleep_ms":
                        e.RequireExtern("declare void @ami_rt_sleep_ms(i64)")
                    case "ami_rt_time_now":
                        e.RequireExtern("declare i64 @ami_rt_time_now()")
                    case "ami_rt_time_add":
                        e.RequireExtern("declare i64 @ami_rt_time_add(i64, i64)")
                    case "ami_rt_time_delta":
                        e.RequireExtern("declare i64 @ami_rt_time_delta(i64, i64)")
                    case "ami_rt_time_unix":
                        e.RequireExtern("declare i64 @ami_rt_time_unix(i64)")
                    case "ami_rt_time_unix_nano":
                        e.RequireExtern("declare i64 @ami_rt_time_unix_nano(i64)")
                    case "ami_rt_signal_register":
                        e.RequireExtern("declare void @ami_rt_signal_register(i64, i64)")
                    case "ami_rt_os_signal_enable":
                        e.RequireExtern("declare void @ami_rt_os_signal_enable(i64)")
                    case "ami_rt_os_signal_disable":
                        e.RequireExtern("declare void @ami_rt_os_signal_disable(i64)")
                    case "ami_rt_install_handler_thunk":
                        e.RequireExtern("declare void @ami_rt_install_handler_thunk(i64, ptr)")
                    case "ami_rt_get_handler_thunk":
                        e.RequireExtern("declare ptr @ami_rt_get_handler_thunk(i64)")
                    case "ami_rt_posix_install_trampoline":
                        e.RequireExtern("declare void @ami_rt_posix_install_trampoline(i64)")
                    case "ami_rt_gpu_blocking_submit":
                        e.RequireExtern("declare ptr @ami_rt_gpu_blocking_submit(ptr)")
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
                        case "ami_rt_owned_ptr":
                            e.RequireExtern("declare ptr @ami_rt_owned_ptr(ptr)")
                        case "ami_rt_owned_new":
                            e.RequireExtern("declare ptr @ami_rt_owned_new(i8*, i64)")
                        case "ami_rt_zeroize_owned":
                            e.RequireExtern("declare void @ami_rt_zeroize_owned(ptr)")
                        case "ami_rt_sleep_ms":
                            e.RequireExtern("declare void @ami_rt_sleep_ms(i64)")
                        case "ami_rt_time_now":
                            e.RequireExtern("declare i64 @ami_rt_time_now()")
                        case "ami_rt_time_add":
                            e.RequireExtern("declare i64 @ami_rt_time_add(i64, i64)")
                        case "ami_rt_time_delta":
                            e.RequireExtern("declare i64 @ami_rt_time_delta(i64, i64)")
                        case "ami_rt_time_unix":
                            e.RequireExtern("declare i64 @ami_rt_time_unix(i64)")
                        case "ami_rt_time_unix_nano":
                            e.RequireExtern("declare i64 @ami_rt_time_unix_nano(i64)")
                        case "ami_rt_signal_register":
                            e.RequireExtern("declare void @ami_rt_signal_register(i64, i64)")
                        case "ami_rt_install_handler_thunk":
                            e.RequireExtern("declare void @ami_rt_install_handler_thunk(i64, ptr)")
                        case "ami_rt_get_handler_thunk":
                            e.RequireExtern("declare ptr @ami_rt_get_handler_thunk(i64)")
                        case "ami_rt_posix_install_trampoline":
                            e.RequireExtern("declare void @ami_rt_posix_install_trampoline(i64)")
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
