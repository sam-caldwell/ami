package logger

import (
    "bytes"
    "testing"
)

func TestStderrSink_WritesToWriter(t *testing.T) {
    var buf bytes.Buffer
    s := NewStderrSink(&buf)
    if err := s.Start(); err != nil { t.Fatalf("start: %v", err) }
    if err := s.Write([]byte("err\n")); err != nil { t.Fatalf("write: %v", err) }
    if buf.String() != "err\n" { t.Fatalf("unexpected: %q", buf.String()) }
    _ = s.Close()
}

