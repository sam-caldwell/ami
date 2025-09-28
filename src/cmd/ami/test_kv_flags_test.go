package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

// CLI flags --kv-metrics and --kv-dump cause kvstore artifacts to be written under build/test/kv/.
func TestTestCmd_KVFlags_EmitArtifacts(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "kv_flags")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    c := newTestCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--kv-metrics", "--kv-dump", dir})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    if _, err := os.Stat(filepath.Join(dir, "build", "test", "kv", "metrics.json")); err != nil { t.Fatalf("metrics missing: %v", err) }
    if _, err := os.Stat(filepath.Join(dir, "build", "test", "kv", "dump.json")); err != nil { t.Fatalf("dump missing: %v", err) }
}

