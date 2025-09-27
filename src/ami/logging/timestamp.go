package logging

import (
    "time"
)

// iso8601UTCms returns ISO-8601 UTC timestamp with millisecond precision.
func iso8601UTCms(t time.Time) string {
    // Always convert to UTC and format with milliseconds and trailing Z.
    return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

