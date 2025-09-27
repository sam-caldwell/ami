package ir

import (
    "encoding/json"
    "testing"
)

func TestEncode_ExprCall_IncludesCallee(t *testing.T) {
    temp := Value{ID: "t1", Type: "any"}
    call := Expr{Op: "call", Callee: "Foo", Args: []Value{{ID: "a", Type: "int"}}, Result: &temp}
    f := Function{Name: "F", Blocks: []Block{{Name: "entry", Instr: []Instruction{call}}}}
    m := Module{Package: "p", Functions: []Function{f}}
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    // quick presence check: walk
    fns := obj["functions"].([]any)
    fn := fns[0].(map[string]any)
    blks := fn["blocks"].([]any)
    blk := blks[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    in := instrs[0].(map[string]any)
    if in["op"] != "EXPR" { t.Fatalf("op: %v", in["op"]) }
    expr := in["expr"].(map[string]any)
    if expr["op"] != "call" || expr["callee"] != "Foo" { t.Fatalf("expr: %v", expr) }
}

