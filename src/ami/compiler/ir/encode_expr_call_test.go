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

func TestEncode_ExprCall_ArgsIncludeTypes(t *testing.T) {
    temp := Value{ID: "t1", Type: "int"}
    call := Expr{Op: "call", Callee: "Foo", Args: []Value{{ID: "a", Type: "string"}}, Result: &temp}
    f := Function{Name: "F", Blocks: []Block{{Name: "entry", Instr: []Instruction{call}}}}
    m := Module{Package: "p", Functions: []Function{f}}
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    fn := fns[0].(map[string]any)
    blks := fn["blocks"].([]any)
    blk := blks[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    in := instrs[0].(map[string]any)
    expr := in["expr"].(map[string]any)
    args := expr["args"].([]any)
    a0 := args[0].(map[string]any)
    if a0["type"] != "string" { t.Fatalf("arg type: %v", a0["type"]) }
}

func TestEncode_ExprCall_SignatureTypes(t *testing.T) {
    temp := Value{ID: "t1", Type: "int"}
    call := Expr{Op: "call", Callee: "Foo", Args: []Value{{ID: "a", Type: "string"}, {ID: "b", Type: "int"}}, Result: &temp}
    f := Function{Name: "F", Blocks: []Block{{Name: "entry", Instr: []Instruction{call}}}}
    m := Module{Package: "p", Functions: []Function{f}}
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    fn := fns[0].(map[string]any)
    blks := fn["blocks"].([]any)
    blk := blks[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    in := instrs[0].(map[string]any)
    expr := in["expr"].(map[string]any)
    at := expr["argTypes"].([]any)
    rt := expr["retTypes"].([]any)
    if len(at) != 2 || at[0] != "string" || at[1] != "int" { t.Fatalf("argTypes: %+v", at) }
    if len(rt) != 1 || rt[0] != "int" { t.Fatalf("retTypes: %+v", rt) }
}
