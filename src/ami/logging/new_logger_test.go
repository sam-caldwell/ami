package logging

import "testing"

func TestNewLogger_FilePair(t *testing.T) {
    if lg, err := New(Options{}); err != nil || lg == nil { t.Fatalf("New: %v %v", lg, err) }
    _ = lg.Close()
}

