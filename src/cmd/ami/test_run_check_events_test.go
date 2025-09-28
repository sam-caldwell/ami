package main

import (
    "bytes"
    "testing"
)

func TestRunCheckEvents_Valid_NoOutput(t *testing.T) {
    var buf bytes.Buffer
    if err := runCheckEvents(&buf); err != nil {
        t.Fatalf("runCheckEvents error: %v", err)
    }
    if buf.Len() != 0 {
        t.Fatalf("expected no output on success; got %q", buf.String())
    }
}

