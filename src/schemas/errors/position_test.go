package errors

import (
    "encoding/json"
    "testing"
)

func TestPosition_JSONKeys(t *testing.T) {
    p := Position{Line: 1, Column: 2, Offset: 3}
    b, err := json.Marshal(p)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    if !contains2(s, "\"line\":1") || !contains2(s, "\"column\":2") || !contains2(s, "\"offset\":3") {
        t.Fatalf("unexpected json: %s", s)
    }
}

// local helpers to avoid name collision with error_test.go
func contains2(s, sub string) bool { return len(s) >= len(sub) && indexOf2(s, sub) >= 0 }
func indexOf2(s, sub string) int {
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub { return i }
    }
    return -1
}
