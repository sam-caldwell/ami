package main

import (
    "bytes"
    "strings"
    "testing"
)

// Verify error path emits JSON with error key and returns a non-nil error.
func TestModAudit_Run_JSON_Error(t *testing.T) {
    var out bytes.Buffer
    // Use a non-existent directory to force an error
    err := runModAudit(&out, "./definitely/missing/workspace", true)
    if err == nil {
        t.Fatalf("expected error on missing workspace")
    }
    s := out.String()
    if !strings.Contains(s, "\"error\"") {
        t.Fatalf("expected JSON error output; got: %s", s)
    }
}

