package ir

import "testing"

func TestDefer_StructAndKind(t *testing.T) {
    d := Defer{Expr: Expr{Op: "call", Callee: "F"}}
    if d.Expr.Op != "call" || d.Expr.Callee != "F" { t.Fatalf("unexpected: %+v", d) }
    if d.isInstruction() != OpDefer { t.Fatalf("kind: %v", d.isInstruction()) }
}

