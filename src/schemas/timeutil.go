package schemas

import (
    "time"
)

// FormatTimestamp returns ISO-8601 UTC timestamp with millisecond precision, e.g., 2025-09-24T17:05:06.123Z
func FormatTimestamp(t time.Time) string {
    return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
