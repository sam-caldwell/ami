package main

import (
    "bytes"
    "testing"
)

func TestPipeline_HelpStubs_Run(t *testing.T) {
    cases := [][]string{
        {"pipeline", "--help"},
        {"pipeline", "visualize", "--help"},
        {"pipeline", "stats", "--help"},
    }
    for _, args := range cases {
        c := newRootCmd()
        var out bytes.Buffer
        c.SetOut(&out)
        c.SetArgs(args)
        if err := c.Execute(); err != nil {
            t.Fatalf("execute %v: %v", args, err)
        }
        if out.Len() == 0 { t.Fatalf("expected help output for %v", args) }
    }
}
