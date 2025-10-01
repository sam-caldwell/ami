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
                // Math multi-result helpers (aggregate returns)
                if ex.Callee == "ami_rt_math_sincos" {
                    e.RequireExtern("declare { double, double } @ami_rt_math_sincos(double)")
                }
                if ex.Callee == "ami_rt_math_frexp" {
                    e.RequireExtern("declare { double, i64 } @ami_rt_math_frexp(double)")
                }
                if ex.Callee == "ami_rt_math_modf" {
                    e.RequireExtern("declare { double, double } @ami_rt_math_modf(double)")
                }
                if ex.Callee == "ami_rt_math_pow10" {
                    e.RequireExtern("declare double @ami_rt_math_pow10(i64)")
                }
                if ex.Callee == "ami_rt_math_inf" {
                    e.RequireExtern("declare double @ami_rt_math_inf(i64)")
                }
                if ex.Callee == "ami_rt_math_isnan" {
                    e.RequireExtern("declare i1 @ami_rt_math_isnan(double)")
                }
                if ex.Callee == "ami_rt_math_isinf" {
                    e.RequireExtern("declare i1 @ami_rt_math_isinf(double, i64)")
                }
                if ex.Callee == "ami_rt_math_signbit" {
                    e.RequireExtern("declare i1 @ami_rt_math_signbit(double)")
                }
                if ex.Callee == "ami_rt_math_nan" {
                    e.RequireExtern("declare double @ami_rt_math_nan()")
                }
                if ex.Callee == "ami_rt_math_remainder" {
                    e.RequireExtern("declare double @ami_rt_math_remainder(double, double)")
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
                    case "ami_rt_metal_available":
                        e.RequireExtern("declare i1 @ami_rt_metal_available()")
                    case "ami_rt_metal_devices":
                        e.RequireExtern("declare ptr @ami_rt_metal_devices()")
                    case "ami_rt_metal_ctx_create":
                        e.RequireExtern("declare ptr @ami_rt_metal_ctx_create(ptr)")
                    case "ami_rt_metal_ctx_destroy":
                        e.RequireExtern("declare void @ami_rt_metal_ctx_destroy(ptr)")
                    case "ami_rt_metal_lib_compile":
                        e.RequireExtern("declare ptr @ami_rt_metal_lib_compile(ptr)")
                    case "ami_rt_metal_pipe_create":
                        e.RequireExtern("declare ptr @ami_rt_metal_pipe_create(ptr, ptr)")
                    case "ami_rt_metal_alloc":
                        e.RequireExtern("declare ptr @ami_rt_metal_alloc(i64)")
                    case "ami_rt_metal_free":
                        e.RequireExtern("declare void @ami_rt_metal_free(ptr)")
                    case "ami_rt_metal_copy_to_device":
                        e.RequireExtern("declare void @ami_rt_metal_copy_to_device(ptr, ptr, i64)")
                    case "ami_rt_metal_copy_from_device":
                        e.RequireExtern("declare void @ami_rt_metal_copy_from_device(ptr, ptr, i64)")
                    case "ami_rt_metal_dispatch_blocking":
                        e.RequireExtern("declare ptr @ami_rt_metal_dispatch_blocking(ptr, ptr, i64, i64, i64, i64, i64, i64)")
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
