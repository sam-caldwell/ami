package ir

import "testing"

func TestInstruction_Kinds(t *testing.T) {
    var i Instruction
    i = Var{}
    if i.isInstruction() != OpVar { t.Fatalf("Var kind") }
    i = Assign{}
    if i.isInstruction() != OpAssign { t.Fatalf("Assign kind") }
    i = Return{}
    if i.isInstruction() != OpReturn { t.Fatalf("Return kind") }
    i = Defer{}
    if i.isInstruction() != OpDefer { t.Fatalf("Defer kind") }
    i = Expr{}
    if i.isInstruction() != OpExpr { t.Fatalf("Expr kind") }
}

