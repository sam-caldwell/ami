package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

// When ami.manifest exists and disagrees with ami.sum, emit E_INTEGRITY_MANIFEST and exit with integrity error.
func TestBuild_ManifestMismatch_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "build_manifest", "mismatch")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // sum: lib@v1.0.0
    sum := []byte(`{ "schema": "ami.sum/v1", "packages": { "lib": { "v1.0.0": "deadbeef" } } }`)
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }
    // manifest: empty packages
    mani := []byte(`{ "schema": "ami.manifest/v1", "packages": {} }`)
    if err := os.WriteFile(filepath.Join(dir, "ami.manifest"), mani, 0o644); err != nil { t.Fatalf("write mani: %v", err) }
    // workspace
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    var buf bytes.Buffer
    err := runBuild(&buf, dir, true, false)
    if err == nil { t.Fatalf("expected integrity error") }
    var m map[string]any
    if e := json.Unmarshal(buf.Bytes(), &m); e != nil { t.Fatalf("json: %v; %s", e, buf.String()) }
    if m["code"] != "E_INTEGRITY_MANIFEST" { t.Fatalf("expected E_INTEGRITY_MANIFEST; got %v", m["code"]) }
}

