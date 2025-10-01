package ir

import (
    "encoding/json"
    "testing"
)

// Verify EncodeModule encodes a RETURN with multiple values and includes types.
func TestEncode_Return_MultiValue_JSONShape(t *testing.T) {
    // Build a minimal function with one block containing a multi-value return.
    f := Function{
        Name: "Pair",
        Blocks: []Block{{
            Name: "entry",
            Instr: []Instruction{
                Return{Values: []Value{{ID: "a", Type: "int"}, {ID: "b", Type: "string"}}},
            },
        }},
    }
    m := Module{Package: "p", Functions: []Function{f}}
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    if len(fns) != 1 { t.Fatalf("functions length: %d", len(fns)) }
    fn := fns[0].(map[string]any)
    blks := fn["blocks"].([]any)
    if len(blks) != 1 { t.Fatalf("blocks length: %d", len(blks)) }
    instrs := blks[0].(map[string]any)["instrs"].([]any)
    if len(instrs) == 0 { t.Fatalf("no instrs") }
    // Find RETURN and assert two values with expected types
    var found bool
    for _, iv := range instrs {
        in := iv.(map[string]any)
        if in["op"].(string) != "RETURN" { continue }
        vals := in["values"].([]any)
        if len(vals) != 2 { t.Fatalf("return values len: %d", len(vals)) }
        t0 := vals[0].(map[string]any)["type"].(string)
        t1 := vals[1].(map[string]any)["type"].(string)
        if t0 != "int" || t1 != "string" {
            t.Fatalf("return value types: %s,%s", t0, t1)
        }
        found = true
    }
    if !found { t.Fatalf("RETURN not found") }
}

