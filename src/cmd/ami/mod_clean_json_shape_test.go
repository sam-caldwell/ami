package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModClean_JSON_Shape(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_clean", "json_shape")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", dir)
    var buf bytes.Buffer
    if err := runModClean(&buf, true); err != nil { t.Fatalf("runModClean: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(buf.Bytes(), &m); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    // Required fields
    if _, ok := m["path"]; !ok { t.Fatalf("missing path") }
    if _, ok := m["removed"]; !ok { t.Fatalf("missing removed") }
    if _, ok := m["created"]; !ok { t.Fatalf("missing created") }
}

