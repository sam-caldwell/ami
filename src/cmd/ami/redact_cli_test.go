package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Verifies that --redact masks fields in debug logs when --verbose is on.
func TestRoot_RedactFlag_RedactsInitTargetInDebugLog(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_redact")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    if err := os.Chdir(dir); err != nil { t.Fatalf("chdir: %v", err) }

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--verbose", "--redact", "target", "init", "--json", "--force"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    closeRootLogger() // ensure file closed for reading

    data, err := os.ReadFile(filepath.Join("build", "debug", "activity.log"))
    if err != nil { t.Fatalf("read activity.log: %v", err) }
    s := string(data)
    if !strings.Contains(s, `"message":"init.target_ready"`) {
        t.Fatalf("expected init.target_ready in debug log: %s", s)
    }
    if !strings.Contains(s, `"target":"[REDACTED]"`) {
        t.Fatalf("expected target redacted; got: %s", s)
    }
    if strings.Contains(s, `"target":"./build"`) {
        t.Fatalf("target value should not appear unredacted: %s", s)
    }
}

