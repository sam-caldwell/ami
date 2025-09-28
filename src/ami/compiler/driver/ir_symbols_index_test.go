package driver

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestCollectExterns_FromIR_AllocAndPanic(t *testing.T) {
    m := ir.Module{Package: "app"}
    fn := ir.Function{Name: "F"}
    blk := ir.Block{Name: "entry"}
    // emit an alloc and a panic op
    blk.Instr = append(blk.Instr, ir.Expr{Op: "alloc", Args: []ir.Value{{ID: "n", Type: "int"}}})
    blk.Instr = append(blk.Instr, ir.Expr{Op: "panic", Args: []ir.Value{{ID: "c", Type: "int"}}})
    fn.Blocks = append(fn.Blocks, blk)
    m.Functions = append(m.Functions, fn)

    ex := collectExterns(m)
    if len(ex) == 0 { t.Fatalf("expected externs from IR ops") }
    hasAlloc, hasPanic := false, false
    for _, s := range ex { if s == "ami_rt_alloc" { hasAlloc = true }; if s == "ami_rt_panic" { hasPanic = true } }
    if !hasAlloc || !hasPanic { t.Fatalf("missing externs: alloc=%v panic=%v set=%v", hasAlloc, hasPanic, ex) }
}

