package ir

import "testing"

func Test_instrsToJSON_CoversAll(t *testing.T) {
    ins := []Instruction{
        Var{Name: "x", Type: "int", Init: &Value{ID: "c1", Type: "int"}, Result: Value{ID: "x", Type: "int"}},
        Assign{DestID: "x", Src: Value{ID: "y", Type: "int"}},
        Return{Values: []Value{{ID: "x", Type: "int"}}},
        Defer{Expr: Expr{Op: "call", Callee: "foo", Args: []Value{{ID: "a", Type: "int"}}}},
        Expr{Op: "call", Callee: "bar", Args: []Value{{ID: "b", Type: "int"}}, Result: &Value{ID: "r", Type: "int"}},
        Phi{Result: Value{ID: "p", Type: "int"}, Incomings: []PhiIncoming{{Value: Value{ID: "v1", Type: "int"}, Label: "L1"}}},
        CondBr{Cond: Value{ID: "c", Type: "bool"}, TrueLabel: "T", FalseLabel: "F"},
        Loop{Name: "loop"},
        Goto{Label: "L1"},
        SetPC{PC: 3},
        Dispatch{Label: "D1"},
        PushFrame{Fn: "f"},
        PopFrame{},
    }
    out := instrsToJSON(ins)
    if len(out) != len(ins) { t.Fatalf("len=%d", len(out)) }
}

