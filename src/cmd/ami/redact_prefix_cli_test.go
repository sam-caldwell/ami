package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Verifies that --redact-prefix masks fields by key prefix in debug logs when --verbose is on.
func TestRoot_RedactPrefixFlag_RedactsInitRootInDebugLog(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_redact_prefix")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    if err := os.Chdir(dir); err != nil { t.Fatalf("chdir: %v", err) }

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    // redact any field whose key starts with "ro" (e.g., "root")
    c.SetArgs([]string{"--verbose", "--redact-prefix", "ro", "init", "--json", "--force"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    closeRootLogger() // ensure debug file flushed

    data, err := os.ReadFile(filepath.Join("build", "debug", "activity.log"))
    if err != nil { t.Fatalf("read activity.log: %v", err) }
    s := string(data)
    if !strings.Contains(s, `"message":"init.pkgroot_ready"`) {
        t.Fatalf("expected init.pkgroot_ready in debug log: %s", s)
    }
    if !strings.Contains(s, `"root":"[REDACTED]"`) {
        t.Fatalf("expected root redacted; got: %s", s)
    }
    if strings.Contains(s, `"root":"./src"`) {
        t.Fatalf("root value should not appear unredacted: %s", s)
    }
}

