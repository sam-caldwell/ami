package main

import (
    "bytes"
    "testing"
)

func TestErrorsSchemaPrint_HiddenCommand_Works(t *testing.T) {
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetErr(&out)
    c.SetArgs([]string{"errors", "schema", "--print"})
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v; out=%s", err, out.String())
    }
    s := out.String()
    if len(s) == 0 || s[0] != '{' {
        t.Fatalf("expected JSON schema output, got: %q", s)
    }
}

