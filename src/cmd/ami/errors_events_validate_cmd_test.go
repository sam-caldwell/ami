package main

import "testing"

func TestErrorsValidateCmd_RequiresFile(t *testing.T) {
    c := newErrorsValidateCmd()
    c.SetArgs([]string{})
    if err := c.Execute(); err == nil {
        t.Fatal("expected error when --file is missing")
    }
}

func TestEventsValidateCmd_RequiresFile(t *testing.T) {
    c := newEventsValidateCmd()
    c.SetArgs([]string{})
    if err := c.Execute(); err == nil {
        t.Fatal("expected error when --file is missing")
    }
}

