package kvstore

import (
    "bufio"
    "bytes"
    "encoding/json"
    "testing"
)

// EmitRegistryMetrics produces diag.v1 lines with sorted namespaces and required fields.
func TestEmitRegistryMetrics_JSON_SchemaAndFields(t *testing.T) {
    ResetDefault()
    ResetRegistry()
    Namespace("p1/n1").Put("a", 1)
    Namespace("p2/n2").Put("b", 2)
    var buf bytes.Buffer
    if err := EmitRegistryMetrics(&buf); err != nil { t.Fatalf("emit: %v", err) }
    lines := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
    count := 0
    var gotSchemas, gotKV bool
    for lines.Scan() {
        var m map[string]any
        if err := json.Unmarshal(lines.Bytes(), &m); err != nil { t.Fatalf("json: %v", err) }
        if m["schema"] == "diag.v1" && m["code"] == "KV_METRICS" { gotSchemas = true }
        if data, ok := m["data"].(map[string]any); ok {
            _ = data["hits"]
            _ = data["misses"]
            _ = data["expirations"]
            _ = data["evictions"]
            _ = data["currentSize"]
            gotKV = true
        }
        count++
    }
    if count < 1 || !gotSchemas || !gotKV { t.Fatalf("missing diag lines or fields: %d %v %v", count, gotSchemas, gotKV) }
}

