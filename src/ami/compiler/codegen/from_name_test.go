package codegen

import "testing"

func TestFromName_Empty(t *testing.T) {
    if b, ok := FromName(""); !ok || b == nil { t.Fatalf("expected default backend") }
}

