package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure EmitModuleLLVMForTarget collects externs for OS signal hooks and emits calls.
func TestEmitTarget_Externs_SignalOSHooks(t *testing.T) {
    m := ir.Module{Package: "app"}
    f := ir.Function{Name: "F"}
    f.Blocks = append(f.Blocks, ir.Block{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "call", Callee: "ami_rt_os_signal_enable", Args: []ir.Value{{ID: "#15", Type: "int64"}}},
        ir.Expr{Op: "call", Callee: "ami_rt_os_signal_disable", Args: []ir.Value{{ID: "#15", Type: "int64"}}},
        ir.Return{},
    }})
    m.Functions = append(m.Functions, f)
    out, err := EmitModuleLLVMForTarget(m, DefaultTriple)
    if err != nil { t.Fatalf("emit: %v", err) }
    wants := []string{
        "declare void @ami_rt_os_signal_enable(i64)",
        "declare void @ami_rt_os_signal_disable(i64)",
        "call void @ami_rt_os_signal_enable(i64 15)",
        "call void @ami_rt_os_signal_disable(i64 15)",
    }
    for _, w := range wants {
        if !strings.Contains(out, w) { t.Fatalf("missing %q in:\n%s", w, out) }
    }
}

