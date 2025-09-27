package logger

import (
    "os"
    "path/filepath"
    "testing"
)

func TestFileSink_AppendsLines(t *testing.T) {
    base := filepath.Join("build", "test", "stdlib_logger")
    if err := os.MkdirAll(base, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    path := filepath.Join(base, "out.log")
    _ = os.Remove(path)
    s := NewFileSink(path, 0o644)
    if err := s.Start(); err != nil { t.Fatalf("start: %v", err) }
    if err := s.Write([]byte("line1\n")); err != nil { t.Fatalf("write1: %v", err) }
    if err := s.Write([]byte("line2\n")); err != nil { t.Fatalf("write2: %v", err) }
    if err := s.Close(); err != nil { t.Fatalf("close: %v", err) }
    data, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    if string(data) != "line1\nline2\n" { t.Fatalf("unexpected: %q", string(data)) }
}

