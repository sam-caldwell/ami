package main

import (
    "bytes"
    "testing"
)

func Test_runCmd_exists(t *testing.T) {
    // Construct command and execute without required flags to hit early error path
    cmd := newRunCmd()
    var out bytes.Buffer
    cmd.SetOut(&out)
    cmd.SetErr(&out)
    // No flags set; RunE should report missing required flags
    if err := cmd.RunE(cmd, nil); err == nil {
        t.Fatalf("expected error for missing --package/--pipeline")
    }
}
