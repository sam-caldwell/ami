package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Verify human output summarizes when ami.sum is missing.
func TestModAudit_Run_Human_NoSum(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_mod_audit", "human_nosum")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0', 'modB 1.0.0' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    var out bytes.Buffer
    if err := runModAudit(&out, dir, false); err != nil { t.Fatalf("run: %v", err) }
    s := out.String()
    if !strings.Contains(s, "ami.sum: not found") {
        t.Fatalf("expected no-sum notice; out=%s", s)
    }
    if !strings.Contains(s, "missing in sum:") {
        t.Fatalf("expected missing-in-sum line; out=%s", s)
    }
}

