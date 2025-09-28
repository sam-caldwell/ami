package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

// When an .ami file has no test pragmas, a default case asserts parse_ok.
func TestRunTest_DefaultAmiCase_ParseOK(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_testcmd", "default_case")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // A simple .ami with no pragmas
    if err := os.WriteFile(filepath.Join(dir, "main.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runTest(&buf, dir, true, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    // The summary (last line) should show ami_tests=1 and ami_failures=0
    lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
    var last map[string]any
    if err := json.Unmarshal(lines[len(lines)-1], &last); err != nil { t.Fatalf("json: %v; %s", err, buf.String()) }
    if int(last["ami_tests"].(float64)) != 1 || int(last["ami_failures"].(float64)) != 0 {
        t.Fatalf("unexpected ami summary: %v", last)
    }
}

