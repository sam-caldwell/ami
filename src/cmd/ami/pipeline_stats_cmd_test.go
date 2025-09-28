package main

import (
    "bytes"
    "encoding/json"
    "testing"
)

func TestPipeline_Stats_JSON_NoVerbose(t *testing.T) {
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json", "pipeline", "stats"})
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    var m map[string]any
    if err := json.Unmarshal(out.Bytes(), &m); err != nil {
        t.Fatalf("json: %v: %q", err, out.String())
    }
    if m["schema"] != "pipeline.stats.v1" { t.Fatalf("schema mismatch: %v", m["schema"]) }
    if act, _ := m["active"].(bool); act {
        t.Fatalf("expected inactive pipeline without --verbose")
    }
}

func TestPipeline_Stats_JSON_Verbose(t *testing.T) {
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"--json", "--verbose", "pipeline", "stats"})
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    var m map[string]any
    if err := json.Unmarshal(out.Bytes(), &m); err != nil {
        t.Fatalf("json: %v: %q", err, out.String())
    }
    if m["schema"] != "pipeline.stats.v1" { t.Fatalf("schema mismatch: %v", m["schema"]) }
    if act, _ := m["active"].(bool); !act {
        t.Fatalf("expected active pipeline with --verbose")
    }
}

