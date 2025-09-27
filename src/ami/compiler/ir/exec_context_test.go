package ir

import (
    "encoding/json"
    "testing"
)

func TestExecContext_JSON_Encode_OmitsWhenNil(t *testing.T) {
    m := Module{Package: "app"}
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if _, ok := obj["execContext"]; ok { t.Fatalf("execContext should be omitted when nil") }
}

func TestExecContext_JSON_Encode_IncludedWhenSet(t *testing.T) {
    m := Module{Package: "app", ExecContext: &ExecContext{Sandbox: true, Limits: map[string]int64{"cpu": 1}, Env: map[string]string{"zone": "A"}}}
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    ec, ok := obj["execContext"].(map[string]any)
    if !ok || len(ec) == 0 { t.Fatalf("execContext missing: %v", obj) }
    if ec["sandbox"] != true { t.Fatalf("sandbox: %v", ec["sandbox"]) }
}

