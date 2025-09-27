package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModGet_CLI_JSON_LocalPath(t *testing.T) {
    wsdir := filepath.Join("build", "test", "mod_get", "cli")
    if err := os.MkdirAll(filepath.Join(wsdir, ".git"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // workspace with util package
    ws := []byte("---\nversion: 1.0.0\npackages:\n  - util:\n      name: util\n      version: 1.0.0\n      root: ./util\n      import: []\n")
    if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.MkdirAll(filepath.Join(wsdir, "util"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(wsdir, "util", "a.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // cache
    cache := filepath.Join("build", "test", "mod_get", "cache_cli")
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    // run via root
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    if err := os.Chdir(wsdir); err != nil { t.Fatalf("chdir: %v", err) }
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"mod", "get", "./util", "--json"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    var res modGetResult
    if err := json.Unmarshal(out.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, out.String()) }
    if res.Name != "util" || res.Version != "1.0.0" { t.Fatalf("unexpected: %+v", res) }
}
