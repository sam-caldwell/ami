package main

import (
    "bytes"
    "encoding/json"
    "testing"
)

func TestVersionCommand_JSON_PrintsObject(t *testing.T) {
    old := version
    version = "v1.2.3-json"
    defer func() { version = old }()

    c := newVersionCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    // Simulate persistent flag present
    c.Flags().Bool("json", true, "")
    _ = c.Flags().Set("json", "true")
    if err := c.RunE(c, nil); err != nil { t.Fatalf("run: %v", err) }
    var m map[string]string
    if err := json.Unmarshal(out.Bytes(), &m); err != nil { t.Fatalf("json: %v; %s", err, out.String()) }
    if m["version"] != "v1.2.3-json" { t.Fatalf("version mismatch: %v", m) }
}

