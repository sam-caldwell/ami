package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

// Ensure ami test accepts common flags: --timeout, --parallel, --failfast, --run.
func TestNewTestCmd_CLI_CommonFlagsAcceptance(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_test_flags_acceptance")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestAlpha(t *testing.T){ }
func TestBeta(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    c := newTestCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json", "--timeout", "1000", "--parallel", "2", "--failfast", "--run", "Alpha", dir})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v\n%s", err, out.String()) }
    // Final JSON summary should be present
    if !bytes.Contains(out.Bytes(), []byte(`"ok":`)) {
        t.Fatalf("missing JSON summary in output: %s", out.String())
    }
}

