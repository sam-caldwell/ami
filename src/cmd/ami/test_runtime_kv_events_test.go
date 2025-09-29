package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestRuntime_KV_Events_JSON_Stream(t *testing.T) {
    dir := t.TempDir()
    content := `#pragma test:case c1
#pragma test:runtime input={}
#pragma test:kv ns="p1/n1" put="x=1" get="x" emit=false
`
    if err := os.WriteFile(filepath.Join(dir, "rt_kv_events_test.ami"), []byte(content), 0o644); err != nil { t.Fatal(err) }
    c := newTestCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json", "--kv-events", dir})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    // Scan NDJSON for KV_METRICS and KV_DUMP diag lines
    dec := json.NewDecoder(bytes.NewReader(out.Bytes()))
    sawMetrics := false
    sawDump := false
    for dec.More() {
        var m map[string]any
        if dec.Decode(&m) != nil { break }
        if code, _ := m["code"].(string); code == "KV_METRICS" { sawMetrics = true }
        if code, _ := m["code"].(string); code == "KV_DUMP" { sawDump = true }
    }
    if !sawMetrics || !sawDump { t.Fatalf("expected KV_METRICS and KV_DUMP in JSON stream; out=\n%s", out.String()) }
}

