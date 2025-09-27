package main

import (
    "bytes"
    "testing"
)

func TestVersionCommand_PrintsVersion(t *testing.T) {
    // Override version for test to ensure wiring uses the variable and not a constant
    old := version
    version = "v9.9.9-test"
    defer func() { version = old }()

    c := newVersionCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    if got := out.String(); got == "" || !bytes.Contains(out.Bytes(), []byte("ami version v9.9.9-test")) {
        t.Fatalf("unexpected output: %q", got)
    }
}
