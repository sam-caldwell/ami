package logger

import (
    "bytes"
    "testing"
)

func TestStdoutSink_WritesToWriter(t *testing.T) {
    var buf bytes.Buffer
    s := NewStdoutSink(&buf)
    if err := s.Start(); err != nil { t.Fatalf("start: %v", err) }
    if err := s.Write([]byte("hello\n")); err != nil { t.Fatalf("write: %v", err) }
    if buf.String() != "hello\n" { t.Fatalf("unexpected: %q", buf.String()) }
    _ = s.Close()
}

