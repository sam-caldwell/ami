package events

import (
    "encoding/json"
    "strings"
    "testing"
)

func TestEvent_MarshalIncludesSchema(t *testing.T) {
    b, err := json.Marshal(Event{})
    if err != nil { t.Fatalf("marshal: %v", err) }
    if string(b) == "{}" || len(b) == 0 { t.Fatalf("expected schema field, got %s", string(b)) }
    if got := string(b); !strings.HasPrefix(got, "{\"schema\":") { t.Fatalf("schema not first: %s", got) }
}
