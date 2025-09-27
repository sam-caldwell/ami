package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestRoot_ModList_JSON_ShowsEntries(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_list", "cli")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "pkgA", "v1.2.3"), 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    if err := os.WriteFile(filepath.Join(dir, "pkgA", "v1.2.3", "f.txt"), []byte("x"), 0o644); err != nil {
        t.Fatalf("write: %v", err)
    }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", dir)

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"mod", "list", "--json"})
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    var res modListResult
    if err := json.Unmarshal(out.Bytes(), &res); err != nil {
        t.Fatalf("json: %v; out=%s", err, out.String())
    }
    if len(res.Entries) != 1 || res.Entries[0].Name != "pkgA" || res.Entries[0].Version != "v1.2.3" {
        t.Fatalf("unexpected entries: %+v", res.Entries)
    }
}
