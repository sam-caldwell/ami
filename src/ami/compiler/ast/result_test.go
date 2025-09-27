package ast

import "testing"

func TestResult_Struct(t *testing.T) {
    r := Result{Type: "int"}
    if r.Type != "int" { t.Fatalf("unexpected: %+v", r) }
}

