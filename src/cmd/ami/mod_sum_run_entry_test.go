package main

import (
    "bytes"
    "testing"
)

func Test_runModSum_missingFile(t *testing.T) {
    var buf bytes.Buffer
    if err := runModSum(&buf, t.TempDir(), true); err == nil {
        t.Fatalf("expected error when ami.sum missing")
    }
}

