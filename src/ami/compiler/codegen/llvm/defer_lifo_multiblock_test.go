package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure defers scheduled in an earlier block are emitted in LIFO order
// before a return in a later block, with correct branching.
func TestEmitter_Defer_LIFO_MultiBlock(t *testing.T) {
    m := ir.Module{Package: "app"}
    fn := ir.Function{Name: "F"}

    // entry block: schedule two defers, then branch to exit
    b0 := ir.Block{Name: "entry"}
    b0.Instr = append(b0.Instr,
        ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{{ID: "h", Type: "Owned"}}}},
        ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{{ID: "h", Type: "Owned"}}}},
        ir.Goto{Label: "exit"},
    )

    // exit block: return
    b1 := ir.Block{Name: "exit"}
    b1.Instr = append(b1.Instr, ir.Return{})

    fn.Blocks = []ir.Block{b0, b1}
    m.Functions = []ir.Function{fn}

    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }

    // Check branch emitted from entry to exit
    if !strings.Contains(out, "entry:\n  br label %exit\n") {
        t.Fatalf("missing branch to exit:\n%s", out)
    }
    // Defers must appear in LIFO order before the return in exit block
    idxRet := strings.Index(out, "\n  ret ")
    if idxRet < 0 { t.Fatalf("no ret in output:\n%s", out) }
    idxB := strings.Index(out, "call void @ami_rt_zeroize_owned")
    idxA := strings.Index(out, "call i64 @ami_rt_owned_len")
    if !(idxB >= 0 && idxA >= 0 && idxB < idxA && idxA < idxRet) {
        t.Fatalf("defer order incorrect: B=%d A=%d ret=%d\n%s", idxB, idxA, idxRet, out)
    }
}

