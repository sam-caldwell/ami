package main

import (
    "bytes"
    "testing"
)

func TestPipelineVisualize_JSON_EmitsGraphV1(t *testing.T) {
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"pipeline", "visualize", "--json"})
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    b := out.Bytes()
    if !bytes.Contains(b, []byte(`"schema":"graph.v1"`)) {
        t.Fatalf("missing schema: %s", string(b))
    }
    if !bytes.Contains(b, []byte(`"name":"Placeholder"`)) {
        t.Fatalf("missing placeholder name: %s", string(b))
    }
}

