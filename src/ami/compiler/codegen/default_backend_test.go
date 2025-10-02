package codegen

import "testing"

func TestDefaultBackend(t *testing.T) {
    if DefaultBackend() == nil { t.Fatalf("nil default backend") }
}

