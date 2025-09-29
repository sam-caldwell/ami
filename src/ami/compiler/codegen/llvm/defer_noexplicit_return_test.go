package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// When no explicit return is present, emitter synthesizes a return at function end
// and must emit all deferred calls in LIFO order before that implicit return.
func TestEmitter_Defer_ImplicitReturn_End(t *testing.T) {
    m := ir.Module{Package: "app"}
    fn := ir.Function{Name: "F"}
    b := ir.Block{Name: "entry"}
    // Schedule A then B
    b.Instr = append(b.Instr,
        ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{{ID: "h", Type: "Owned"}}}},
        ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{{ID: "h", Type: "Owned"}}}},
    )
    // No explicit return
    fn.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{fn}

    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    // There must be a function-end ret and defers before it in LIFO order
    idxRet := strings.LastIndex(out, "\n  ret ")
    if idxRet < 0 { t.Fatalf("no synthetic ret in output:\n%s", out) }
    idxB := strings.Index(out, "call void @ami_rt_zeroize_owned")
    idxA := strings.Index(out, "call i64 @ami_rt_owned_len")
    if !(idxB >= 0 && idxA >= 0 && idxB < idxA && idxA < idxRet) {
        t.Fatalf("defer order incorrect: B=%d A=%d ret=%d\n%s", idxB, idxA, idxRet, out)
    }
}

