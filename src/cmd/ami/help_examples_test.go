package main

import (
    "bytes"
    "testing"
)

func TestHelpCommand_IncludesExamples(t *testing.T) {
    root := newRootCmd()
    var out bytes.Buffer
    root.SetOut(&out)
    root.SetArgs([]string{"help"})
    if err := root.Execute(); err != nil {
        t.Fatalf("execute help: %v", err)
    }
    s := out.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) {
        t.Fatalf("expected help to include Examples section; got: %s", s)
    }
    if !bytes.Contains([]byte(s), []byte("ami init")) {
        t.Fatalf("expected example snippet present; got: %s", s)
    }
}

