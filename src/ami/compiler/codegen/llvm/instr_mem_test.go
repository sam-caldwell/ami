package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestLowerExpr_Load_Store_GEP(t *testing.T) {
    m := ir.Module{Package: "app"}
    fn := ir.Function{Name: "F"}
    b := ir.Block{Name: "entry"}
    // All temps
    p := ir.Value{ID: "p", Type: "ptr"}
    idx := ir.Value{ID: "i", Type: "int64"}
    // GEP: q = gep p, i
    q := ir.Value{ID: "q", Type: "ptr"}
    b.Instr = append(b.Instr, ir.Expr{Op: "gep", Args: []ir.Value{p, idx}, Result: &q})
    // load r from q (i64)
    r := ir.Value{ID: "r", Type: "int64"}
    b.Instr = append(b.Instr, ir.Expr{Op: "load", Args: []ir.Value{q}, Result: &r})
    // store r into p
    b.Instr = append(b.Instr, ir.Expr{Op: "store", Args: []ir.Value{r, p}})
    b.Instr = append(b.Instr, ir.Return{})
    fn.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{fn}

    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "%q = getelementptr i8, ptr %p, i64 %i") {
        t.Fatalf("missing gep:\n%s", out)
    }
    if !strings.Contains(out, "%r = load i64, ptr %q") {
        t.Fatalf("missing load:\n%s", out)
    }
    if !strings.Contains(out, "store i64 %r, ptr %p") {
        t.Fatalf("missing store:\n%s", out)
    }
}

