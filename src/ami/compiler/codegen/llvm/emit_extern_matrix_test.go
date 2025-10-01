package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Build a dense extern matrix to exercise extern collection across both Expr and Defer paths
// and across both EmitModuleLLVM and EmitModuleLLVMForTarget.
func TestEmit_Extern_Matrix_AllHooks(t *testing.T) {
    names := []string{
        "ami_rt_zeroize",
        "ami_rt_owned_len",
        "ami_rt_owned_ptr",
        "ami_rt_owned_new",
        "ami_rt_zeroize_owned",
        "ami_rt_sleep_ms",
        "ami_rt_time_now",
        "ami_rt_time_add",
        "ami_rt_time_delta",
        "ami_rt_time_unix",
        "ami_rt_time_unix_nano",
        "ami_rt_signal_register",
        "ami_rt_os_signal_enable",
        "ami_rt_os_signal_disable",
        "ami_rt_install_handler_thunk",
        "ami_rt_get_handler_thunk",
        "ami_rt_posix_install_trampoline",
    }
    // Build function with direct calls
    f1 := ir.Function{Name: "F1"}
    var instr1 []ir.Instruction
    for _, n := range names { instr1 = append(instr1, ir.Expr{Op: "call", Callee: n}) }
    instr1 = append(instr1, ir.Return{})
    f1.Blocks = append(f1.Blocks, ir.Block{Name: "entry", Instr: instr1})
    // Build function with the same calls wrapped in Defer
    f2 := ir.Function{Name: "F2"}
    var instr2 []ir.Instruction
    for _, n := range names { instr2 = append(instr2, ir.Defer{Expr: ir.Expr{Op: "call", Callee: n}}) }
    instr2 = append(instr2, ir.Return{})
    f2.Blocks = append(f2.Blocks, ir.Block{Name: "entry", Instr: instr2})

    m := ir.Module{Package: "app", Functions: []ir.Function{f1, f2}}
    // Emit with default helper
    outA, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit A: %v", err) }
    // Emit for explicit target
    outB, err := EmitModuleLLVMForTarget(m, DefaultTriple)
    if err != nil { t.Fatalf("emit B: %v", err) }
    // A small set of representative externs must appear in both outputs
    must := []string{
        "declare void @ami_rt_zeroize(ptr, i64)",
        "declare i64 @ami_rt_owned_len(ptr)",
        "declare ptr @ami_rt_owned_ptr(ptr)",
        "declare ptr @ami_rt_owned_new(i8*, i64)",
        "declare void @ami_rt_zeroize_owned(ptr)",
        "declare void @ami_rt_sleep_ms(i64)",
        "declare i64 @ami_rt_time_now()",
        "declare i64 @ami_rt_time_add(i64, i64)",
        "declare i64 @ami_rt_time_delta(i64, i64)",
        "declare i64 @ami_rt_time_unix(i64)",
        "declare i64 @ami_rt_time_unix_nano(i64)",
        "declare void @ami_rt_signal_register(i64, i64)",
        "declare void @ami_rt_os_signal_enable(i64)",
        "declare void @ami_rt_os_signal_disable(i64)",
        "declare void @ami_rt_install_handler_thunk(i64, ptr)",
        "declare ptr @ami_rt_get_handler_thunk(i64)",
        "declare void @ami_rt_posix_install_trampoline(i64)",
    }
    for _, w := range must {
        if !strings.Contains(outA, w) { t.Fatalf("missing extern in A: %s\n%s", w, outA) }
        if !strings.Contains(outB, w) { t.Fatalf("missing extern in B: %s\n%s", w, outB) }
    }
}

