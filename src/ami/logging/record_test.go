package logging

import (
    "encoding/json"
    "regexp"
    "strings"
    "testing"
    "time"
)

func TestRecordMarshalJSON_DeterministicOrdering(t *testing.T) {
    r := Record{
        Timestamp: time.Date(2025, 9, 24, 17, 5, 6, 123000000, time.UTC),
        Level:     LevelInfo,
        Message:   "hello\r\nworld",
        Package:   "ami/logging",
        Fields: map[string]any{
            "b": 2,
            "a": 1,
        },
    }
    b, err := json.Marshal(r)
    if err != nil {
        t.Fatalf("marshal error: %v", err)
    }
    s := string(b)
    // Expected order and CRLF normalized
    if !strings.Contains(s, `"timestamp":"2025-09-24T17:05:06.123Z"`) {
        t.Fatalf("missing/invalid timestamp: %s", s)
    }
    want := `{"timestamp":"2025-09-24T17:05:06.123Z","level":"info","package":"ami/logging","message":"hello\nworld","fields":{"a":1,"b":2}}`
    if s != want {
        t.Fatalf("unexpected json order/body:\nwant: %s\n got: %s", want, s)
    }
}

func TestJSONFormatter_NoColorSequences(t *testing.T) {
    r := Record{Timestamp: time.Unix(0, 0), Level: LevelError, Message: "boom"}
    out := JSONFormatter{}.Format(r)
    // No ANSI escape sequences in JSON
    if ansiPattern().Match(out) {
        t.Fatalf("JSON output should not contain ANSI codes: %q", string(out))
    }
}

func ansiPattern() *regexp.Regexp {
    return regexp.MustCompile(`\x1b\[[0-9;]*m`)
}
