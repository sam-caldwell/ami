package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure EmitModuleLLVMForTarget collects math externs like EmitModuleLLVM.
func TestEmitModuleLLVMForTarget_MathExterns(t *testing.T) {
    m := ir.Module{Package: "app"}
    f := ir.Function{Name: "F"}
    // Call sincos runtime helper with aggregate return
    f.Blocks = []ir.Block{{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "call", Callee: "ami_rt_math_sincos", Args: []ir.Value{{ID: "x", Type: "float64"}}, Results: []ir.Value{{ID: "s", Type: "float64"}, {ID: "c", Type: "float64"}}},
        ir.Return{},
    }}}
    m.Functions = []ir.Function{f}
    out, err := EmitModuleLLVMForTarget(m, DefaultTriple)
    if err != nil { t.Fatalf("emit: %v", err) }
    want := "declare { double, double } @ami_rt_math_sincos(double)"
    if !strings.Contains(out, want) {
        t.Fatalf("missing math extern in target emission: %q\n%s", want, out)
    }
}

