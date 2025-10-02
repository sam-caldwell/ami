package llvm

import (
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

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
                    // Math externs (mirror of EmitModuleLLVM)
                    case "ami_rt_math_sincos":
                        e.RequireExtern("declare { double, double } @ami_rt_math_sincos(double)")
                    case "ami_rt_math_frexp":
                        e.RequireExtern("declare { double, i64 } @ami_rt_math_frexp(double)")
                    case "ami_rt_math_modf":
                        e.RequireExtern("declare { double, double } @ami_rt_math_modf(double)")
                    case "ami_rt_math_pow10":
                        e.RequireExtern("declare double @ami_rt_math_pow10(i64)")
                    case "ami_rt_math_inf":
                        e.RequireExtern("declare double @ami_rt_math_inf(i64)")
                    case "ami_rt_math_isnan":
                        e.RequireExtern("declare i1 @ami_rt_math_isnan(double)")
                    case "ami_rt_math_isinf":
                        e.RequireExtern("declare i1 @ami_rt_math_isinf(double, i64)")
                    case "ami_rt_math_signbit":
                        e.RequireExtern("declare i1 @ami_rt_math_signbit(double)")
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

