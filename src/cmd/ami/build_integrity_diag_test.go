package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Test that runBuild emits per-item E_INTEGRITY diagnostics in JSON mode
// when ami.sum entries are missing from the cache.
func TestBuild_PerItemIntegrityDiagnostics_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "build_integrity", "per_item")
    sumPath := filepath.Join(dir, "ami.sum")
    wsPath := filepath.Join(dir, "ami.workspace")
    cache := filepath.Join(dir, "cache")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }

    // Minimal workspace with one package that imports a remote dep "example.org/lib@^1.0.0"
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'example.org/lib@^1.0.0' ]\n")
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.WriteFile(wsPath, ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // Sum declares example.org/lib at v1.0.0, but cache is empty => missingInCache
    sum := []byte(`{ "schema": "ami.sum/v1", "packages": { "example.org/lib": { "v1.0.0": "deadbeef" } } }`)
    if err := os.WriteFile(sumPath, sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }

    // Point AMI_PACKAGE_CACHE to our empty cache
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)

    var buf bytes.Buffer
    err := runBuild(&buf, dir, true, false)
    if err == nil { t.Fatalf("expected integrity error") }

    // Decode NDJSON records and look for a per-item E_INTEGRITY for missingInCache
    dec := json.NewDecoder(&buf)
    var sawItem, sawSummary bool
    for dec.More() {
        var r diag.Record
        if derr := dec.Decode(&r); derr != nil { t.Fatalf("decode: %v; out=%s", derr, buf.String()) }
        if r.Code == "E_INTEGRITY" && r.File == "ami.sum" && r.Data != nil {
            if kind, ok := r.Data["kind"].(string); ok && kind == "missingInCache" {
                sawItem = true
            }
        }
        if r.Code == "E_INTEGRITY" && r.File == "ami.workspace" {
            sawSummary = true
        }
    }
    if !sawItem { t.Fatalf("expected per-item missingInCache diagnostic; got: %s", buf.String()) }
    if !sawSummary { t.Fatalf("expected summary E_INTEGRITY record; got: %s", buf.String()) }
}

