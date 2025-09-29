package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Human mode prints per-test concise lines and per-package case counts.
func TestRunTest_Human_PerTestLines_And_PackageCases(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "human_per_test")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
func TestB(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    if err := runTest(&out, dir, false, false, 0); err != nil { t.Fatalf("runTest: %v\n%s", err, out.String()) }
    s := out.String()
    if !(strings.Contains(s, "test: example.com/tmp TestA pass") || strings.Contains(s, "test: tmp TestA pass")) {
        t.Fatalf("missing per-test line for TestA: %s", s)
    }
    if !(strings.Contains(s, "test: example.com/tmp TestB pass") || strings.Contains(s, "test: tmp TestB pass")) {
        t.Fatalf("missing per-test line for TestB: %s", s)
    }
    if !strings.Contains(s, "cases=2") {
        t.Fatalf("missing per-package cases count: %s", s)
    }
}

