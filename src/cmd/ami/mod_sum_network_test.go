package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

// When a sum entry references an unreachable git source and cache is missing,
// mod sum should report integrity failure (missing) rather than hang.
func TestModSum_UnreachableGitSource_ReportsIntegrity_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "network_fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // workspace minimal
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // sum with unreachable git source
    sum := []byte(`{"schema":"ami.sum/v1","packages":{"repo":{"version":"v0.1.0","sha256":"","source":"git+ssh://git@invalid.invalid/org/repo.git"}}}`)
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }
    var buf bytes.Buffer
    err := runModSum(&buf, dir, true)
    if err == nil { t.Fatalf("expected integrity error due to missing cache and unreachable source") }
    if exit.UnwrapCode(err) != exit.Integrity { t.Fatalf("expected Integrity exit; got %v", exit.UnwrapCode(err)) }
    var res modSumResult
    if e := json.Unmarshal(buf.Bytes(), &res); e != nil { t.Fatalf("json: %v; out=%s", e, buf.String()) }
    if res.Ok { t.Fatalf("expected ok=false; res=%+v", res) }
    if len(res.Missing) == 0 { t.Fatalf("expected missing to be non-empty") }
}

