package logger

import (
    "encoding/json"
    "testing"
)

func TestStdJSONUnmarshal_Behavior(t *testing.T) {
    in := []byte(`{"a":1,"b":"x"}`)
    var m map[string]any
    if err := stdJSONUnmarshal(in, &m); err != nil {
        t.Fatalf("stdJSONUnmarshal: %v", err)
    }
    // Cross-check with encoding/json directly for parity
    var m2 map[string]any
    if err := json.Unmarshal(in, &m2); err != nil { t.Fatalf("json.Unmarshal: %v", err) }
    if len(m) != len(m2) || m["a"].(float64) != 1 || m["b"].(string) != "x" {
        t.Fatalf("unexpected map: %#v", m)
    }
}

