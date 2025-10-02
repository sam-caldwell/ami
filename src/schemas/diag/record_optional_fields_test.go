package diag

import (
    "encoding/json"
    "strings"
    "testing"
    "time"
)

func TestDiagRecord_OptionalFieldsOmitted(t *testing.T) {
    r := Record{Timestamp: time.Unix(0, 0), Level: Info, Code: "I000", Message: "ok"}
    b, _ := json.Marshal(r)
    s := string(b)
    if strings.Contains(s, "package") || strings.Contains(s, "file") || strings.Contains(s, "pos") || strings.Contains(s, "endPos") || strings.Contains(s, "data") {
        t.Fatalf("optional fields should not appear when empty: %s", s)
    }
}

