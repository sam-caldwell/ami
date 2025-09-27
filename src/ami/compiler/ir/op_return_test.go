package ir

import "testing"

func TestReturn_StructAndKind(t *testing.T) {
    r := Return{Values: []Value{{ID: "r0", Type: "int"}}}
    if len(r.Values) != 1 || r.Values[0].Type != "int" { t.Fatalf("unexpected: %+v", r) }
    if r.isInstruction() != OpReturn { t.Fatalf("kind: %v", r.isInstruction()) }
}

