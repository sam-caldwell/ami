package ir

import "testing"

func TestVar_StructAndKind(t *testing.T) {
    v := Var{Name: "x", Type: "int", Init: &Value{ID: "t0", Type: "int"}, Result: Value{ID: "x0", Type: "int"}}
    if v.Name != "x" || v.Type != "int" || v.Init == nil || v.Result.ID != "x0" { t.Fatalf("unexpected: %+v", v) }
    if v.isInstruction() != OpVar { t.Fatalf("kind: %v", v.isInstruction()) }
}

