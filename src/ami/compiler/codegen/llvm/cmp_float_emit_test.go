package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Verify floating-point comparators emit fcmp with ordered predicates.
func TestLowerExpr_Cmp_Float(t *testing.T) {
    m := ir.Module{Package: "app"}
    fn := ir.Function{Name: "F"}
    b := ir.Block{Name: "entry"}
    a := ir.Value{ID: "a", Type: "float64"}
    c := ir.Value{ID: "c", Type: "float64"}
    r := ir.Value{ID: "r", Type: "bool"}
    b.Instr = append(b.Instr, ir.Expr{Op: "eq", Args: []ir.Value{a, c}, Result: &r})
    b.Instr = append(b.Instr, ir.Return{})
    fn.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{fn}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "%r = fcmp oeq double %a, %c") {
        t.Fatalf("expected fcmp oeq for double eq; got:\n%s", out)
    }
}

