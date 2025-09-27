package ast

import "testing"

func TestParam_Struct(t *testing.T) {
    p := Param{Name: "a", Type: "int"}
    if p.Name != "a" || p.Type != "int" { t.Fatalf("unexpected: %+v", p) }
}

