package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func Test_lowerVar_FormatsComment(t *testing.T) {
    out := lowerVar(ir.Var{Name: "x", Type: "int", Result: ir.Value{ID: "x0", Type: "int"}})
    if !strings.Contains(out, "; var x : i64 as %x0") { t.Fatalf("unexpected: %s", out) }
}
