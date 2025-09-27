package diag

import (
    "encoding/json"
    "testing"
    "strings"
)

func TestLevel_Constants(t *testing.T) {
    if Info != Level("info") || Warn != Level("warn") || Error != Level("error") {
        t.Fatalf("level constants mismatch: %q %q %q", Info, Warn, Error)
    }
}

func TestPosition_JSONFields(t *testing.T) {
    p := Position{Line: 10, Column: 20, Offset: 30}
    b, err := json.Marshal(p)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    if !(strings.Contains(s, "\"line\":10") && strings.Contains(s, "\"column\":20") && strings.Contains(s, "\"offset\":30")) {
        t.Fatalf("unexpected position json: %s", s)
    }
}

