package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Ensure JSON output includes per-package summary events (ami.test.pkg.v1).
func TestRunTest_JSON_IncludesPerPackageSummaryEvents(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "json_pkg_events")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    s := buf.String()
    if !strings.Contains(s, `"schema":"ami.test.pkg.v1"`) {
        t.Fatalf("missing per-package summary events in JSON output: %s", s)
    }
}

// Tolerates an empty package tree (no tests) without error and emits OK in human mode.
func TestRunTest_EmptyPackageTree_OK(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "empty")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Minimal module, but no *_test.go files
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    var out bytes.Buffer
    if err := runTest(&out, dir, false, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    if !strings.Contains(out.String(), "test: OK") {
        t.Fatalf("expected human OK summary for empty package tree; got: %s", out.String())
    }
}

