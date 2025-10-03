package errorpipe

import (
    "bytes"
    "encoding/json"
    "errors"
    "strings"
    "testing"
)

// TestWrite_JSON_NoData ensures Write emits a proper errors.v1 line without data when nil/empty.
func TestWrite_JSON_NoData(t *testing.T) {
    var buf bytes.Buffer
    if err := Write(&buf, "E_TEST", "boom", "t.ami", nil); err != nil { t.Fatalf("Write: %v", err) }
    out := buf.String()
    if !strings.HasSuffix(out, "\n") { t.Fatalf("missing newline: %q", out) }
    line := strings.TrimSuffix(out, "\n")
    var m map[string]any
    if err := json.Unmarshal([]byte(line), &m); err != nil { t.Fatalf("json: %v", err) }
    if m["schema"] != "errors.v1" { t.Fatalf("schema: %v", m["schema"]) }
    if m["level"] != "error" { t.Fatalf("level: %v", m["level"]) }
    if m["code"] != "E_TEST" { t.Fatalf("code: %v", m["code"]) }
    if m["message"] != "boom" { t.Fatalf("message: %v", m["message"]) }
    if m["file"] != "t.ami" { t.Fatalf("file: %v", m["file"]) }
    if _, ok := m["data"]; ok { t.Fatalf("unexpected data present: %v", m["data"]) }
}

// TestWrite_JSON_WithData ensures Write includes data when provided.
func TestWrite_JSON_WithData(t *testing.T) {
    var buf bytes.Buffer
    if err := Write(&buf, "E_TEST2", "boom2", "f.ami", map[string]any{"a": "b", "n": 3}); err != nil { t.Fatalf("Write: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m); err != nil { t.Fatalf("json: %v", err) }
    d, ok := m["data"].(map[string]any)
    if !ok { t.Fatalf("expected data map; m=%v", m) }
    if d["a"] != "b" || int(d["n"].(float64)) != 3 { t.Fatalf("bad data: %v", d) }
}

// errWriter implements io.Writer and always fails.
type errWriter struct{ err error }
func (e errWriter) Write(p []byte) (int, error) { return 0, e.err }

// TestWrite_WriteError bubbles up underlying writer failures.
func TestWrite_WriteError(t *testing.T) {
    ew := errWriter{err: errors.New("sink fail")}
    if err := Write(ew, "E", "m", "f", map[string]any{"k":"v"}); err == nil { t.Fatalf("expected error from writer") }
}

// TestWrite_UnsupportedData_MarshalError ensures unsupported JSON types cause an error.
func TestWrite_UnsupportedData_MarshalError(t *testing.T) {
    var buf bytes.Buffer
    ch := make(chan int)
    err := Write(&buf, "E_JSON", "bad", "f", map[string]any{"ch": ch})
    if err == nil { t.Fatalf("expected marshal error for unsupported type") }
    if buf.Len() != 0 { t.Fatalf("expected no bytes written on error; got %d", buf.Len()) }
}
