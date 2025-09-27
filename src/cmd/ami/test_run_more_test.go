package main

import (
    "bytes"
    "encoding/json"
    "io"
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

// RunTest JSON summary counts across multiple subpackages.
func TestRunTest_JSON_MultiPackageCounts(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "json_multi")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "a"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(dir, "b"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    aSrc := `package a
import "testing"
func TestA(t *testing.T){ }
`
    bSrc := `package b
import "testing"
func TestB(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "a", "a_test.go"), []byte(aSrc), 0o644); err != nil { t.Fatalf("write a: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "b", "b_test.go"), []byte(bSrc), 0o644); err != nil { t.Fatalf("write b: %v", err) }
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, false, 2); err != nil { t.Fatalf("runTest: %v", err) }
    lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
    last := lines[len(lines)-1]
    if !strings.Contains(string(last), `"packages":2`) || !strings.Contains(string(last), `"tests":2`) || !strings.Contains(string(last), `"failures":0`) {
        t.Fatalf("unexpected summary counts: %s", string(last))
    }
}

// AMI directives: assert parse_fail with position and count.
func TestRunTest_AMI_Directives_ParseFail_PositionAndCount(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "ami_parsefail_pos")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    // Missing 'package' should produce a parse error near start of file
    src := "#pragma test:case pos\n#pragma test:assert parse_fail msg=\"expected 'package'\" count=1 line=1 column=1 offset=0\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, false, 0); err != nil { t.Fatalf("runTest: %v\n%s", err, buf.String()) }
}

// AMI directives: mismatch should return an error from runTest.
func TestRunTest_AMI_Directives_Mismatch_ReturnsError(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "ami_mismatch")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    // Expect parse_fail but file is valid → mismatch should produce amiFailures > 0 and error
    src := "package app\n#pragma test:case bad\n#pragma test:assert parse_fail\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    if err := runTest(&out, dir, false, false, 0); err == nil {
        t.Fatalf("expected error from mismatched AMI directive; out=%s", out.String())
    }
}

// CLI wiring: `ami test` respects --json and --verbose flags and writes artifacts.
func TestNewTestCmd_CLI_JSON_Verbose(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_test_cmd")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    _ = os.Chdir(dir)
    c := newTestCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json", "--verbose"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    // JSON summary present
    if !strings.Contains(out.String(), `"ok":`) { t.Fatalf("missing JSON summary: %s", out.String()) }
    // Artifacts present (check relative to cwd, which is already dir)
    logPath := filepath.Join("build", "test", "test.log")
    manPath := filepath.Join("build", "test", "test.manifest")
    if _, err := os.Stat(logPath); err != nil { t.Fatalf("test.log missing: %v", err) }
    if _, err := os.Stat(manPath); err != nil { t.Fatalf("test.manifest missing: %v", err) }
    // test.log should contain go test JSON events
    if b, err := os.ReadFile(logPath); err == nil {
        if !strings.Contains(string(b), `"Action":"run"`) {
            t.Fatalf("expected JSON events in test.log; got: %s", string(b))
        }
    }
}

// CLI: ensure --packages flag is accepted and run succeeds.
func TestNewTestCmd_CLI_PackagesFlag(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_test_cmd_pkgs")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    _ = os.Chdir(dir)
    c := newTestCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json", "--packages", "2"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    if !strings.Contains(out.String(), `"ok":`) { t.Fatalf("missing JSON summary: %s", out.String()) }
}

// CLI: `ami test [path]` runs tests in the given directory.
func TestNewTestCmd_CLI_PathArgument(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_test_cmd_path")
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
    c.SetArgs([]string{"--json", dir})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    if !strings.Contains(out.String(), `"ok":`) { t.Fatalf("missing JSON summary: %s", out.String()) }
}

// JSON stream includes final ami_tests and ami_failures fields.
func TestRunTest_JSON_IncludesAmiSummaryFields(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "json_ami_fields")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    // One AMI directive test case that passes
    src := "package app\n#pragma test:case c1\n#pragma test:assert parse_ok\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
    last := lines[len(lines)-1]
    s := string(last)
    if !strings.Contains(s, `"ami_tests":`) || !strings.Contains(s, `"ami_failures":`) {
        t.Fatalf("missing ami summary fields: %s", s)
    }
}

// JSON stream contains ami.test.v1 events with expected fields.
func TestRunTest_JSON_AmiEventsContainFields(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "json_ami_events_fields")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    src := "package app\n#pragma test:case case1\n#pragma test:assert parse_ok\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
    found := false
    for _, ln := range lines[:len(lines)-1] { // exclude final summary
        s := string(ln)
        if strings.Contains(s, `"schema":"ami.test.v1"`) && strings.Contains(s, `"file":"main.ami"`) && strings.Contains(s, `"case":"case1"`) {
            found = true
            break
        }
    }
    if !found { t.Fatalf("ami.test.v1 event missing or malformed: %s", buf.String()) }
}

// Human mode emits AMI summary line and OK when only AMI cases pass.
func TestRunTest_Human_AmiSummaryLine(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "human_ami_summary")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    src := "package app\n#pragma test:case c1\n#pragma test:assert parse_ok\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    if err := runTest(&out, dir, false, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    s := out.String()
    if !strings.Contains(s, "test: ami ok=1 fail=0") || !strings.Contains(s, "test: OK") {
        t.Fatalf("missing human AMI summary/OK: %s", s)
    }
}

// Failing go tests should write messages to stderr.
// Note: stderr forwarding on failures depends on whether `go test -json` writes errors to stderr.
// We rely on higher-level CLI tests elsewhere; avoiding brittle stderr assertions here.
// JSON mode with failing tests returns error and ok=false summary.
func TestRunTest_JSON_FailingSummaryAndError(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "json_fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestFail(t *testing.T){ t.Fatal("boom") }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    err := runTest(&buf, dir, true, false, 0)
    if err == nil { t.Fatalf("expected error from failing tests") }
    // Parse summary
    lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
    var last map[string]any
    if e := json.Unmarshal(lines[len(lines)-1], &last); e != nil { t.Fatalf("summary json: %v", e) }
    if ok, _ := last["ok"].(bool); ok { t.Fatalf("expected ok=false, got ok=true: %v", last) }
    if int(last["failures"].(float64)) < 1 { t.Fatalf("expected failures>=1: %v", last) }
}

// Verbose manifest includes AMI entries sorted deterministically.
func TestRunTest_Verbose_AmiManifestOrder(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "ami_manifest_order")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    // Two AMI files with cases out of lexical order
    a := "package app\n#pragma test:case zeta\n#pragma test:assert parse_ok\nfunc F(){}\n"
    b := "package app\n#pragma test:case alpha\n#pragma test:assert parse_ok\nfunc G(){}\n"
    if err := os.WriteFile(filepath.Join(dir, "b.ami"), []byte(b), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "a.ami"), []byte(a), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runTest(io.Discard, dir, false, true, 0); err != nil { t.Fatalf("runTest: %v", err) }
    bts, err := os.ReadFile(filepath.Join(dir, "build", "test", "test.manifest"))
    if err != nil { t.Fatalf("read manifest: %v", err) }
    lines := strings.Split(strings.TrimSpace(string(bts)), "\n")
    var amis []string
    for _, ln := range lines {
        if strings.HasPrefix(ln, "ami:") { amis = append(amis, ln) }
    }
    if len(amis) != 2 { t.Fatalf("expected two ami entries, got: %v", amis) }
    // Expect a.ami before b.ami deterministically
    if !(strings.Contains(amis[0], "ami:a.ami") && strings.Contains(amis[1], "ami:b.ami")) {
        t.Fatalf("AMI entries not sorted by filename: %v", amis)
    }
}
