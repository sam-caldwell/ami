package main

import (
    "bytes"
    "testing"
)

func TestRoot_Help_IncludesExamples(t *testing.T) {
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{})
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    s := out.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) {
        t.Fatalf("expected root help examples; got: %s", s)
    }
    if !bytes.Contains([]byte(s), []byte("ami init")) {
        t.Fatalf("expected root example snippet; got: %s", s)
    }
}

