package errorpipe

import (
    "encoding/json"
    "os"
    "testing"
)

// TestDefault_WritesJSONToStderr verifies Default writes a single errors.v1 line to stderr.
func TestDefault_WritesJSONToStderr(t *testing.T) {
    r, w, err := os.Pipe()
    if err != nil { t.Fatalf("pipe: %v", err) }
    old := os.Stderr
    os.Stderr = w
    defer func(){ os.Stderr = old }()

    if err := Default("E_DEF", "default path", "t.ami", map[string]any{"case": "c1"}); err != nil {
        t.Fatalf("Default: %v", err)
    }
    // Close writer side to allow read to complete
    _ = w.Close()
    // Read content
    var buf [4096]byte
    n, _ := r.Read(buf[:])
    if n == 0 { t.Fatalf("no bytes written to stderr") }
    b := buf[:n]
    if b[len(b)-1] != '\n' { t.Fatalf("missing trailing newline: %q", string(b)) }
    line := b[:len(b)-1]
    var m map[string]any
    if err := json.Unmarshal(line, &m); err != nil { t.Fatalf("json: %v", err) }
    if m["schema"] != "errors.v1" { t.Fatalf("want schema errors.v1, got %v", m["schema"]) }
    if m["level"] != "error" { t.Fatalf("want level=error, got %v", m["level"]) }
    if m["code"] != "E_DEF" { t.Fatalf("want code=E_DEF, got %v", m["code"]) }
    if m["message"] != "default path" { t.Fatalf("want message, got %v", m["message"]) }
    if m["file"] != "t.ami" { t.Fatalf("want file=t.ami, got %v", m["file"]) }
    if _, ok := m["timestamp"].(string); !ok { t.Fatalf("timestamp missing or not string: %v", m["timestamp"]) }
    // data should exist and contain case
    data, ok := m["data"].(map[string]any)
    if !ok { t.Fatalf("expected data map present; m=%v", m) }
    if data["case"] != "c1" { t.Fatalf("expected data.case=c1; got %v", data["case"]) }
}
