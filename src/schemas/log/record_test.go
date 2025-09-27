package logschema

import (
    "encoding/json"
    "testing"
    "time"
)

func TestLogRecord_JSONOrderingAndOptionalFields(t *testing.T) {
    r := Record{
        Timestamp: time.Date(2025, 9, 24, 17, 5, 6, 123000000, time.UTC),
        Level:     Info,
        Message:   "hello",
        Package:   "pkg/mod",
        Fields:    map[string]any{"b": 2, "a": 1},
        Pipeline:  "stdout",
        Node:      "worker-1",
    }
    b, err := json.Marshal(r)
    if err != nil {
        t.Fatalf("marshal: %v", err)
    }
    want := `{"schema":"log.v1","timestamp":"2025-09-24T17:05:06.123Z","level":"info","package":"pkg/mod","message":"hello","fields":{"a":1,"b":2},"pipeline":"stdout","node":"worker-1"}`
    if string(b) != want {
        t.Fatalf("unexpected json order/body:\nwant: %s\n got: %s", want, string(b))
    }
}

