package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// TestLowerExpr_EdgeCases covers less-traveled branches in lowerExpr for coverage:
// - logical and with non-bool result type (logic-nonbool)
// - band with float result type (band-float)
// - bor with float result type (bor-float)
func TestLowerExpr_EdgeCases(t *testing.T) {
    m := ir.Module{Package: "p"}
    f := ir.Function{Name: "F"}
    b := ir.Block{Name: "entry"}
    // logic-nonbool: result declared as i64 via Type "int"
    b.Instr = append(b.Instr, ir.Expr{Op: "and", Args: []ir.Value{{ID: "a", Type: "bool"}, {ID: "b", Type: "bool"}}, Result: &ir.Value{ID: "r0", Type: "int"}})
    // band-float and bor-float
    b.Instr = append(b.Instr, ir.Expr{Op: "band", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "r1", Type: "float64"}})
    b.Instr = append(b.Instr, ir.Expr{Op: "bor", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "r2", Type: "float64"}})
    f.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{f}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "; expr logic-nonbool") {
        t.Fatalf("missing logic-nonbool comment:\n%s", out)
    }
    if !strings.Contains(out, "; expr band-float") {
        t.Fatalf("missing band-float comment:\n%s", out)
    }
    if !strings.Contains(out, "; expr bor-float") {
        t.Fatalf("missing bor-float comment:\n%s", out)
    }
}

