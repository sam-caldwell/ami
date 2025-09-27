package ir

import "testing"

func TestBlock_Struct(t *testing.T) {
    b := Block{Name: "entry"}
    if b.Name != "entry" || len(b.Instr) != 0 { t.Fatalf("unexpected: %+v", b) }
}

