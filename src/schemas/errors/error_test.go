package errors

import (
    "encoding/json"
    "testing"
    "time"
)

func TestError_JSONStableAndValidate(t *testing.T) {
    e := Error{
        Timestamp: time.Unix(1700000000, 0).UTC(),
        Level:   "error",
        Code:    "E_SAMPLE",
        Message: "sample",
        File:    "main.ami",
        Pos:     &Position{Line: 1, Column: 2, Offset: 3},
        Data:    map[string]any{"k": 1},
    }
    b, err := json.Marshal(e)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    // schema should be first and code/level/message present
    if want := "{" + "\"schema\":\"errors.v1\""; len(s) < len(want) || s[:len(want)] != want {
        t.Fatalf("schema first: %s", s)
    }
    if !contains(s, "\"code\":\"E_SAMPLE\"") || !contains(s, "\"level\":\"error\"") || !contains(s, "\"message\":\"sample\"") {
        t.Fatalf("fields missing: %s", s)
    }
    if err := Validate(e); err != nil { t.Fatalf("validate: %v", err) }
}

func contains(s, sub string) bool { return len(s) >= len(sub) && indexOf(s, sub) >= 0 }

func indexOf(s, sub string) int {
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub { return i }
    }
    return -1
}
