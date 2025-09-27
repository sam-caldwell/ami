package main

import (
    "bytes"
    "testing"
)

func TestExec_StubCommands_RunHelp(t *testing.T) {
    cases := [][]string{
        {"build", "--help"},
        {"test", "--help"},
        {"mod", "get", "--help"},
        {"mod", "list", "--help"},
        {"mod", "sum", "--help"},
    }
    for _, args := range cases {
        c := newRootCmd()
        var out bytes.Buffer
        c.SetOut(&out)
        c.SetArgs(args)
        if err := c.Execute(); err != nil {
            t.Fatalf("execute %v: %v", args, err)
        }
        // Non-error execution is sufficient for stub coverage
    }
}
