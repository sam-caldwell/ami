package ir

import "testing"

func TestExpr_StructAndKind(t *testing.T) {
    e := Expr{Op: "call", Callee: "G", Args: []Value{{ID: "a0", Type: "int"}}}
    r := Value{ID: "t0", Type: "int"}
    e.Result = &r
    e.ParamTypes = []string{"int"}
    e.ParamNames = []string{"a"}
    e.ResultTypes = []string{"int"}
    if e.Op != "call" || e.Callee != "G" || e.Result == nil || e.ParamNames[0] != "a" { t.Fatalf("unexpected: %+v", e) }
    if e.isInstruction() != OpExpr { t.Fatalf("kind: %v", e.isInstruction()) }
}

