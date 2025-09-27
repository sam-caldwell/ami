package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    . "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestModAudit_JSON_Summary(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_audit", "json")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // workspace imports remote modules
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0', 'modB >= 1.0.0' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // sum missing, expect both missingInSum
    var out bytes.Buffer
    if err := runModAudit(&out, dir, true); err != nil { t.Fatalf("runModAudit: %v", err) }
    var res modAuditResult
    if err := json.Unmarshal(out.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, out.String()) }
    if !res.SumFound { /* ok */ } else { t.Fatalf("sumFound should be false") }
    if len(res.MissingInSum) != 2 { t.Fatalf("missingInSum: %v", res.MissingInSum) }
}

func TestModAudit_Human_OKWhenSatisfied(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_audit", "ok")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // workspace
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // sum + cache entry
    cache := filepath.Join(dir, "cache")
    if err := os.MkdirAll(filepath.Join(cache, "modA", "1.2.3"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "modA", "1.2.3", "x.txt"), []byte("hi"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // compute hash with workspace helper
    sha, err := HashDir(filepath.Join(cache, "modA", "1.2.3"))
    if err != nil { t.Fatalf("hash: %v", err) }
    // ami.sum
    sum := Manifest{Schema: "ami.sum/v1"}
    sum.Set("modA", "1.2.3", sha)
    if err := sum.Save(filepath.Join(dir, "ami.sum")); err != nil { t.Fatalf("save sum: %v", err) }
    // set env to point cache for integrity validation
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)

    var out bytes.Buffer
    if err := runModAudit(&out, dir, false); err != nil { t.Fatalf("runModAudit: %v; out=%s", err, out.String()) }
    if !bytes.Contains(out.Bytes(), []byte("ok:")) {
        t.Fatalf("expected ok summary; out=%s", out.String())
    }
}
