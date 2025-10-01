package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure modulo on floating types is not emitted and leaves a deterministic comment.
func TestLowerExpr_Mod_Double_NotSupported(t *testing.T) {
    m := ir.Module{Package: "p"}
    f := ir.Function{Name: "F"}
    b := ir.Block{Name: "entry"}
    a := ir.Value{ID: "a", Type: "float64"}
    d := ir.Value{ID: "d", Type: "float64"}
    r := ir.Value{ID: "r", Type: "float64"}
    b.Instr = append(b.Instr, ir.Expr{Op: "mod", Args: []ir.Value{a, d}, Result: &r})
    b.Instr = append(b.Instr, ir.Return{})
    f.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{f}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "; expr mod-unsupported-double") {
        t.Fatalf("expected mod-unsupported-double comment; got:\n%s", out)
    }
}

