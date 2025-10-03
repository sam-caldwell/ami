package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func Test_lowerAssign_FormatsComment(t *testing.T) {
    out := lowerAssign(ir.Assign{DestID: "x0", Src: ir.Value{ID: "c1", Type: "int"}})
    if !strings.Contains(out, "; assign %x0 = %c1") { t.Fatalf("unexpected: %s", out) }
}
