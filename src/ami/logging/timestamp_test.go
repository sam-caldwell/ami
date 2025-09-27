package logging

import (
    "testing"
    "time"
)

func TestISO8601UTCms_Format(t *testing.T) {
    ts := time.Date(2025, 9, 24, 17, 5, 6, 789000000, time.UTC)
    if s := iso8601UTCms(ts); s != "2025-09-24T17:05:06.789Z" {
        t.Fatalf("unexpected: %s", s)
    }
}
