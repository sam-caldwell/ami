package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure externs for OS signal hooks are recognized from Defer expressions as well as normal calls.
func TestEmit_Externs_From_Defer_SignalHooks(t *testing.T) {
    m := ir.Module{Package: "app"}
    f := ir.Function{Name: "F"}
    // Defer enable and disable calls
    f.Blocks = append(f.Blocks, ir.Block{Name: "entry", Instr: []ir.Instruction{
        ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_os_signal_enable", Args: []ir.Value{{ID: "#2", Type: "int64"}}}},
        ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_os_signal_disable", Args: []ir.Value{{ID: "#2", Type: "int64"}}}},
        ir.Return{},
    }})
    m.Functions = append(m.Functions, f)
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    wants := []string{
        "declare void @ami_rt_os_signal_enable(i64)",
        "declare void @ami_rt_os_signal_disable(i64)",
    }
    for _, w := range wants {
        if !strings.Contains(out, w) { t.Fatalf("missing extern from defer: %s\n%s", w, out) }
    }
}

