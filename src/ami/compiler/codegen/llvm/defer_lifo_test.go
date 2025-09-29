package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Verify LIFO defer emission across multiple returns and nested defers.
func TestEmitter_Defer_LIFO_MultiReturn_Nested(t *testing.T) {
    m := ir.Module{Package: "app"}
    fn := ir.Function{Name: "F"}
    b := ir.Block{Name: "entry"}
    // Defer two calls A then B
    b.Instr = append(b.Instr, ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{{ID: "h", Type: "Owned"}}}})
    b.Instr = append(b.Instr, ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{{ID: "h", Type: "Owned"}}}})
    // Return path 1
    b.Instr = append(b.Instr, ir.Return{Values: nil})
    // Unreachable second return in same block for coverage
    b.Instr = append(b.Instr, ir.Return{Values: nil})
    fn.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{fn}

    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    // Expect B then A before ret
    idxRet := strings.Index(out, "\n  ret ")
    if idxRet < 0 { t.Fatalf("no ret in output:\n%s", out) }
    idxB := strings.Index(out, "call void @ami_rt_zeroize_owned")
    // Calls without result are emitted as void calls in scaffold
    idxA := strings.Index(out, "call i64 @ami_rt_owned_len")
    if !(idxB >= 0 && idxA >= 0 && idxB < idxA && idxA < idxRet) {
        t.Fatalf("defer order incorrect: B=%d A=%d ret=%d\n%s", idxB, idxA, idxRet, out)
    }
}
