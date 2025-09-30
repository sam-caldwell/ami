package ir

import (
    "testing"
)

func TestEncodeModule_IncludesEventMeta(t *testing.T) {
    m := Module{
        Package: "app",
        EventMeta: &EventMeta{Schema: "eventmeta.v1", Fields: []string{"id", "ts", "attempt", "trace"}},
    }
    b, err := EncodeModule(m)
    if err != nil { t.Fatalf("encode error: %v", err) }
    s := string(b)
    if !contains(s, "\"eventmeta\"") || !contains(s, "eventmeta.v1") || !contains(s, "\"attempt\"") {
        t.Fatalf("expected eventmeta fields in JSON, got: %s", s)
    }
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (indexOf(s, sub) >= 0) }

func indexOf(s, sub string) int {
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub { return i }
    }
    return -1
}

