package diag

import (
    "encoding/json"
    "strings"
    "testing"
)

func TestPosition_JSONFields(t *testing.T) {
    p := Position{Line: 10, Column: 20, Offset: 30}
    b, err := json.Marshal(p)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    if !(strings.Contains(s, "\"line\":10") && strings.Contains(s, "\"column\":20") && strings.Contains(s, "\"offset\":30")) {
        t.Fatalf("unexpected position json: %s", s)
    }
}

