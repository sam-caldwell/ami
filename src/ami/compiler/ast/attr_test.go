package ast

import "testing"

func TestAttr_Struct(t *testing.T) {
    a := Attr{Name: "merge.Buffer"}
    if a.Name != "merge.Buffer" { t.Fatalf("unexpected: %+v", a) }
}

