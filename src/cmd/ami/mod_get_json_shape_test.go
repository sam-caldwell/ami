package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModGet_JSON_Shape_Local(t *testing.T) {
    wsdir := filepath.Join("build", "test", "mod_get", "json_shape")
    _ = os.RemoveAll(wsdir)
    if err := os.MkdirAll(filepath.Join(wsdir, "util"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), []byte("version: 1.0.0\npackages:\n  - util:\n      name: util\n      version: 1.0.0\n      root: ./util\n      import: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.WriteFile(filepath.Join(wsdir, "util", "x.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    cache := filepath.Join("build", "test", "mod_get", "json_cache")
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    var buf bytes.Buffer
    if err := runModGet(&buf, wsdir, "./util", true); err != nil { t.Fatalf("runModGet: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(buf.Bytes(), &m); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    for _, k := range []string{"source", "name", "version", "path"} {
        if _, ok := m[k]; !ok { t.Fatalf("missing %s in JSON result", k) }
    }
}

