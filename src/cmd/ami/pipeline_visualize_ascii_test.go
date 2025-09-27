package main

import (
    "bytes"
    "testing"
)

func TestPipelineVisualize_ASCII_RendersLine(t *testing.T) {
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"pipeline", "visualize"})
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    got := out.String()
    want := "[ingress] --> (worker) --> [egress]\n"
    if got != want {
        t.Fatalf("unexpected ascii output\n got: %q\nwant: %q", got, want)
    }
}

