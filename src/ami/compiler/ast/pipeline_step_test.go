package ast

import "testing"

func TestStepStmt_Struct(t *testing.T) {
    s := StepStmt{Name: "Ingress"}
    if s.Name != "Ingress" || len(s.Args) != 0 { t.Fatalf("unexpected: %+v", s) }
}

