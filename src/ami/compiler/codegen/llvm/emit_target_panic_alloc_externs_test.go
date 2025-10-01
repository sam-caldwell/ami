package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure ForTarget collects externs for panic/alloc when specified as op forms.
func TestEmitTarget_Externs_PanicAlloc_OpForms(t *testing.T) {
    m := ir.Module{Package: "app"}
    f := ir.Function{Name: "F"}
    // Use op forms (not callee) to exercise the op-based extern detection paths.
    f.Blocks = append(f.Blocks, ir.Block{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "panic", Args: []ir.Value{{ID: "#1", Type: "int32"}}},
        ir.Expr{Op: "alloc", Args: []ir.Value{{ID: "#64", Type: "int64"}}, Result: &ir.Value{ID: "p", Type: "ptr"}},
        ir.Return{},
    }})
    m.Functions = append(m.Functions, f)
    out, err := EmitModuleLLVMForTarget(m, DefaultTriple)
    if err != nil { t.Fatalf("emit: %v", err) }
    wants := []string{
        "declare void @ami_rt_panic(i32)",
        "declare ptr @ami_rt_alloc(i64)",
    }
    for _, w := range wants {
        if !strings.Contains(out, w) { t.Fatalf("missing %q in:\n%s", w, out) }
    }
}

