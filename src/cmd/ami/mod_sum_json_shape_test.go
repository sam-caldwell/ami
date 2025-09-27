package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModSum_JSON_Shape_Minimal(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "json_shape")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), []byte(`{"schema":"ami.sum/v1","packages":{}}`), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runModSum(&buf, dir, true); err != nil { t.Fatalf("runModSum: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(buf.Bytes(), &m); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    for _, k := range []string{"schema", "packages", "ok"} {
        if _, ok := m[k]; !ok { t.Fatalf("missing %s in JSON result", k) }
    }
}

