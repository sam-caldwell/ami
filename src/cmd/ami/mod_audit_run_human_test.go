package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

// Verify human summary prints ok when requirements are satisfied and present in cache.
func TestModAudit_Run_Human_OK(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_mod_audit", "human_ok")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // workspace with a single requirement
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // Create a cache populated via AMI_PACKAGE_CACHE and a matching ami.sum
    cache := filepath.Join(dir, "cache")
    if err := os.MkdirAll(filepath.Join(cache, "modA", "1.2.3"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "modA", "1.2.3", "x.txt"), []byte("hi"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    sha, err := hashDirLike(cache, filepath.Join("modA", "1.2.3"))
    if err != nil { t.Fatalf("hash: %v", err) }
    sum := []byte("{\n  \"schema\": \"ami.sum/v1\",\n  \"packages\": {\n    \"modA\": {\n      \"1.2.3\": \"" + sha + "\"\n    }\n  }\n}\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }
    // Ensure the environment exposes the cache root for hashing
    t.Setenv("AMI_PACKAGE_CACHE", cache)

    var out bytes.Buffer
    if err := runModAudit(&out, dir, false); err != nil { t.Fatalf("run: %v", err) }
    if !bytes.Contains(out.Bytes(), []byte("ok:")) {
        t.Fatalf("expected ok summary; out=%s", out.String())
    }
}

