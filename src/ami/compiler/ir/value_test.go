package ir

import "testing"

func TestValue_StructFields(t *testing.T) {
    v := Value{ID: "t0", Type: "int"}
    if v.ID != "t0" || v.Type != "int" { t.Fatalf("unexpected: %+v", v) }
}

