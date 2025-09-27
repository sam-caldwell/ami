package main

import (
    "bytes"
    "testing"
)

func TestHelpCommand_PrintsEmbeddedDocs(t *testing.T) {
    root := newRootCmd()
    var out bytes.Buffer
    root.SetOut(&out)
    root.SetErr(&out)
    root.SetArgs([]string{"help"})
    if err := root.Execute(); err != nil {
        t.Fatalf("execute help: %v", err)
    }
    if !bytes.Contains(out.Bytes(), []byte("AMI Help")) {
        t.Fatalf("expected embedded help content, got: %s", out.String())
    }
}

