package logger

import "time"

// FormatTimestamp returns ISO-8601 UTC with milliseconds and 'Z'.
func FormatTimestamp(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z") }

