package logging

import "testing"

func TestNewLogger_FilePair(t *testing.T) {
    lg, err := New(Options{})
    if err != nil || lg == nil { t.Fatalf("New: %v %v", lg, err) }
    _ = lg.Close()
}
