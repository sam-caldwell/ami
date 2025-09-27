package ir

import "testing"

func TestAssign_StructAndKind(t *testing.T) {
    a := Assign{DestID: "x0", Src: Value{ID: "t1", Type: "int"}}
    if a.DestID != "x0" || a.Src.Type != "int" { t.Fatalf("unexpected: %+v", a) }
    if a.isInstruction() != OpAssign { t.Fatalf("kind: %v", a.isInstruction()) }
}

