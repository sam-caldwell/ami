package main

import (
    "bytes"
    "testing"
)

func TestVersionCommand_PrintsVersion(t *testing.T) {
    c := newVersionCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    if got := out.String(); got == "" || !bytes.Contains(out.Bytes(), []byte("ami version ")) {
        t.Fatalf("unexpected output: %q", got)
    }
}

