package ir

import (
    "encoding/json"
    "testing"
)

func TestEncodeModule_SimpleFunction(t *testing.T) {
    // Build a trivial function
    vX := Value{ID: "x", Type: "int"}
    insts := []Instruction{
        Var{Name: "x", Type: "int", Result: vX},
        Assign{DestID: "x", Src: Value{ID: "c1", Type: "int"}},
        Return{Values: []Value{vX}},
    }
    f := Function{Name: "main", Params: nil, Results: []Value{{ID: "", Type: "int"}}, Blocks: []Block{{Name: "entry", Instr: insts}}}
    m := Module{Package: "app", Functions: []Function{f}}

    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    // Validate JSON structure parses and contains expected keys
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v\n%s", err, string(b)) }
    if obj["schema"] != "ir.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    if obj["package"] != "app" { t.Fatalf("package: %v", obj["package"]) }
}

