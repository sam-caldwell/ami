package driver

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure collectExterns captures zeroize + Owned helpers and deferred calls.
func TestCollectExterns_FromIR_Owned_And_Defer(t *testing.T) {
    m := ir.Module{Package: "app"}
    fn := ir.Function{Name: "F"}
    blk := ir.Block{Name: "entry"}
    // Direct calls
    blk.Instr = append(blk.Instr, ir.Expr{Op: "call", Callee: "ami_rt_zeroize", Args: []ir.Value{{ID: "p", Type: "ptr"}, {ID: "n", Type: "int64"}}})
    blk.Instr = append(blk.Instr, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{{ID: "h", Type: "Owned"}}})
    blk.Instr = append(blk.Instr, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{{ID: "h", Type: "Owned"}}})
    blk.Instr = append(blk.Instr, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{{ID: "p", Type: "ptr"}, {ID: "n", Type: "int64"}}})
    // Deferred zeroize-owned
    blk.Instr = append(blk.Instr, ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{{ID: "h", Type: "Owned"}}}})
    fn.Blocks = append(fn.Blocks, blk)
    m.Functions = append(m.Functions, fn)

    ex := collectExterns(m)
    want := map[string]bool{
        "ami_rt_zeroize": true,
        "ami_rt_owned_len": true,
        "ami_rt_owned_ptr": true,
        "ami_rt_owned_new": true,
        "ami_rt_zeroize_owned": true,
    }
    for k := range want {
        found := false
        for _, s := range ex { if s == k { found = true; break } }
        if !found { t.Fatalf("missing extern %s in set %v", k, ex) }
    }
}

