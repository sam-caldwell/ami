package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
    "strings"
)

// Creates a tiny Go module with tests and runs ami test against it in verbose mode.
func TestRunTest_Verbose_WritesLogAndManifest(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "ok")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    // two tests
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
func TestB(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }

    if err := runTest(os.Stdout, dir, false, true); err != nil { t.Fatalf("runTest: %v", err) }
    if _, err := os.Stat(filepath.Join(dir, "build", "test", "test.log")); err != nil { t.Fatalf("test.log missing: %v", err) }
    if _, err := os.Stat(filepath.Join(dir, "build", "test", "test.manifest")); err != nil { t.Fatalf("test.manifest missing: %v", err) }
}

func TestRunTest_Verbose_ManifestContainsTests(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "manifest_contents")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
func TestB(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runTest(os.Stdout, dir, false, true); err != nil { t.Fatalf("runTest: %v", err) }
    b, err := os.ReadFile(filepath.Join(dir, "build", "test", "test.manifest"))
    if err != nil { t.Fatalf("read manifest: %v", err) }
    s := string(b)
    if !strings.Contains(s, "example.com/tmp TestA") || !strings.Contains(s, "example.com/tmp TestB") {
        t.Fatalf("manifest missing tests: %s", s)
    }
}

func TestRunTest_JSONSummary_TrueOnSuccess(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "json_ok")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, false); err != nil { t.Fatalf("runTest: %v", err) }
    if !strings.Contains(buf.String(), `{"ok":true}`) {
        t.Fatalf("expected JSON ok:true, got: %s", buf.String())
    }
}

func TestRunTest_FailingTests_ReturnsErrorAndNoPanic(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestFail(t *testing.T){ t.Fatal("boom") }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runTest(os.Stdout, dir, false, false); err == nil { t.Fatalf("expected error from failing test run") }
}
