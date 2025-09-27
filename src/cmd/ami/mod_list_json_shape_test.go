package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModList_JSON_Shape(t *testing.T) {
    cache := filepath.Join("build", "test", "mod_list", "json_shape")
    _ = os.RemoveAll(cache)
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    var buf bytes.Buffer
    if err := runModList(&buf, true); err != nil { t.Fatalf("runModList: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(buf.Bytes(), &m); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    if _, ok := m["path"]; !ok { t.Fatalf("missing path") }
    if _, ok := m["entries"]; !ok { t.Fatalf("missing entries") }
}

