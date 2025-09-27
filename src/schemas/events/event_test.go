package events

import (
    "encoding/json"
    "strings"
    "testing"
    "time"
)

func TestEvent_MarshalJSON_IncludesSchemaAndFields(t *testing.T) {
    e := Event{
        ID:        "evt-123",
        Timestamp: time.Date(2025, 9, 24, 17, 5, 6, 123000000, time.UTC),
        Attempt:   1,
        Trace:     map[string]any{"trace_id": "abc"},
        Payload:   map[string]any{"k": 1},
    }
    b, err := json.Marshal(e)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    if !strings.Contains(s, "\"schema\":\"events.v1\"") { t.Fatalf("missing schema: %s", s) }
    if !strings.Contains(s, "\"id\":\"evt-123\"") { t.Fatalf("missing id: %s", s) }
    if !strings.Contains(s, "\"timestamp\":\"2025-09-24T17:05:06.000Z\"") { t.Fatalf("bad ts: %s", s) }
}

func TestValidate_HappyAndSadPaths(t *testing.T) {
    good := Event{ID: "x"}
    if err := Validate(good); err != nil { t.Fatalf("good event invalid: %v", err) }
    bad := Event{}
    if err := Validate(bad); err == nil { t.Fatalf("expected error for missing id") }
}

