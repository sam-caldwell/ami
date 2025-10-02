package diag

import (
    "encoding/json"
    "strings"
    "testing"
    "time"
)

func TestRecord_MarshalIncludesSchema(t *testing.T) {
    r := Record{Timestamp: time.Unix(0,0).UTC(), Level: Info}
    b, err := json.Marshal(r)
    if err != nil { t.Fatalf("marshal: %v", err) }
    if got := string(b); !strings.HasPrefix(got, "{\"schema\":") { t.Fatalf("schema not first: %s", got) }
}
