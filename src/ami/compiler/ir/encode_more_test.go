package ir

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestEncodeModule_CoversTopLevelAndInstrs(t *testing.T) {
    // Build a module with various fields set
    m := Module{
        Package:         "app",
        Concurrency:     4,
        Backpressure:    "dropOldest",
        TelemetryEnabled: true,
        Capabilities:    []string{"io", "network"},
        TrustLevel:      "untrusted",
        Directives: []Directive{{Domain: "concurrency", Key: "level", Value: "4", Args: []string{"level"}, Params: map[string]string{"level": "4"}}},
        Functions: []Function{{
            Name:   "F",
            Params: []Value{{ID: "p1", Type: "int"}},
            Results: []Value{{ID: "r1", Type: "int"}},
            Decorators: []Decorator{{Name: "metrics"}},
            Blocks: []Block{{
                Name: "entry",
                Instr: []Instruction{
                    Var{Name: "x", Type: "int", Result: Value{ID: "v1", Type: "int"}},
                    Var{Name: "y", Type: "int", Result: Value{ID: "v2", Type: "int"}, Init: &Value{ID: "v1", Type: "int"}},
                    Assign{DestID: "v1", Src: Value{ID: "v2", Type: "int"}},
                    Expr{Op: "call", Callee: "G", Args: []Value{{ID: "a1", Type: "int"}}, Result: &Value{ID: "c1", Type: "int"}, ParamTypes: []string{"int"}},
                    Return{Values: []Value{{ID: "c1", Type: "int"}}},
                },
            }},
        }},
    }
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["package"] != "app" { t.Fatalf("package: %v", obj["package"]) }
    if obj["concurrency"].(float64) != 4 { t.Fatalf("concurrency: %v", obj["concurrency"]) }
    if obj["backpressurePolicy"].(string) != "dropOldest" { t.Fatalf("backpressurePolicy: %v", obj["backpressurePolicy"]) }
    if obj["telemetryEnabled"].(bool) != true { t.Fatalf("telemetryEnabled: %v", obj["telemetryEnabled"]) }
    if tr, ok := obj["trustLevel"].(string); !ok || tr == "" { t.Fatalf("trustLevel: %v", obj["trustLevel"]) }
    // Inspect function call signature block
    funs := obj["functions"].([]any)
    if len(funs) != 1 { t.Fatalf("functions len: %d", len(funs)) }
    f0 := funs[0].(map[string]any)
    blocks := f0["blocks"].([]any)
    if len(blocks) != 1 { t.Fatalf("blocks len: %d", len(blocks)) }
    instrs := blocks[0].(map[string]any)["instrs"].([]any)
    hasCall := false
    for _, in := range instrs {
        im := in.(map[string]any)
        if im["op"] == "EXPR" {
            e := im["expr"].(map[string]any)
            if e["op"] == "call" {
                if _, ok := e["sig"].(map[string]any); !ok { t.Fatalf("missing sig in call: %v", e) }
                hasCall = true
            }
        }
    }
    if !hasCall { t.Fatalf("call expr not found in instrs: %v", instrs) }
}

func TestWriteDebug_WritesFile(t *testing.T) {
    m := Module{Package: "pkg", Functions: []Function{{Name: "F"}}}
    if err := WriteDebug(m); err != nil { t.Fatalf("WriteDebug: %v", err) }
    p := filepath.Join("build", "debug", "ir", "pkg.json")
    // File existence is enough; content validated in encode test
    if _, err := os.Stat(p); err != nil { t.Fatalf("missing debug file: %v", err) }
}
