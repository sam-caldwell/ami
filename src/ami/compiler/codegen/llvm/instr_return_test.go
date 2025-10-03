package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func Test_lowerReturn_Void(t *testing.T) {
    s, err := lowerReturn(ir.Return{})
    if err != nil || !strings.Contains(s, "ret void") { t.Fatalf("unexpected: %q err=%v", s, err) }
}

func Test_lowerReturn_Single(t *testing.T) {
    s, err := lowerReturn(ir.Return{Values: []ir.Value{{ID: "x", Type: "int"}}})
    if err != nil || !strings.Contains(s, "ret i64 %x") { t.Fatalf("unexpected: %q err=%v", s, err) }
}

func Test_lowerReturn_Multi(t *testing.T) {
    s, err := lowerReturn(ir.Return{Values: []ir.Value{{ID: "a", Type: "int"}, {ID: "b", Type: "int"}, {ID: "c", Type: "int"}}})
    if err != nil { t.Fatalf("err: %v", err) }
    if !strings.Contains(s, "insertvalue {i64, i64, i64} undef, i64 %a, 0") { t.Fatalf("agg start: %s", s) }
    if !strings.Contains(s, "ret {i64, i64, i64} %ret_agg_a2") { t.Fatalf("agg ret: %s", s) }
}
