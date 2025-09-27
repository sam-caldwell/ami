package main

import (
    "bytes"
    "encoding/json"
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

    if err := runTest(os.Stdout, dir, false, true, 0); err != nil { t.Fatalf("runTest: %v", err) }
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
    if err := runTest(os.Stdout, dir, false, true, 0); err != nil { t.Fatalf("runTest: %v", err) }
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
    if err := runTest(&buf, dir, true, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    // Summary should be last line
    lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
    var last map[string]any
    if err := json.Unmarshal(lines[len(lines)-1], &last); err != nil { t.Fatalf("json: %v; %s", err, buf.String()) }
    if ok, _ := last["ok"].(bool); !ok { t.Fatalf("ok=false in summary: %v", last) }
    // Single module/single test
    if int(last["packages"].(float64)) != 1 || int(last["tests"].(float64)) != 1 || int(last["failures"].(float64)) != 0 {
        t.Fatalf("unexpected counts: %v", last)
    }
}

func TestRunTest_JSON_StreamsEventsAndSummary(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "json_stream")
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
    if !strings.Contains(s, `"Action":"run"`) { t.Fatalf("expected streamed events in JSON output: %s", s) }
    lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
    var last map[string]any
    if err := json.Unmarshal(lines[len(lines)-1], &last); err != nil { t.Fatalf("summary json: %v", err) }
    if ok, _ := last["ok"].(bool); !ok { t.Fatalf("expected ok=true summary; last=%v", last) }
    if int(last["packages"].(float64)) != 1 || int(last["tests"].(float64)) != 1 || int(last["failures"].(float64)) != 0 {
        t.Fatalf("unexpected counts: %v", last)
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
    if err := runTest(os.Stdout, dir, false, false, 0); err == nil { t.Fatalf("expected error from failing test run") }
}

func TestRunTest_AMIDirectives_ParseOK_And_Manifest(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "ami_ok")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    // AMI file with parse-ok
    src := "package app\n#pragma test:case parseok\n#pragma test:assert parse_ok\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // Run verbose to get manifest
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, true, 0); err != nil { t.Fatalf("runTest: %v", err) }
    // Manifest should include ami entry
    b, err := os.ReadFile(filepath.Join(dir, "build", "test", "test.manifest"))
    if err != nil { t.Fatalf("read manifest: %v", err) }
    if !strings.Contains(string(b), "ami:main.ami parseok\n") { t.Fatalf("missing ami manifest entry: %s", string(b)) }
}

func TestRunTest_AMIDirectives_ParseFail_TriggersError(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "ami_fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    // AMI with expected parse_fail but actually ok -> should error
    src := "package app\n#pragma test:case bad\n#pragma test:assert parse_fail\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var stdout bytes.Buffer
    if err := runTest(&stdout, dir, false, false, 0); err == nil {
        t.Fatalf("expected error from ami directive failure")
    }
}

func TestRunTest_Human_PerPackageSummary(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "human_pkg_summary")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){}
func TestB(t *testing.T){}
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    if err := runTest(&out, dir, false, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    s := out.String()
    if !strings.Contains(s, "test: pkg example.com/tmp ok=2 fail=0") {
        t.Fatalf("expected per-package summary; got: %s", s)
    }
}
