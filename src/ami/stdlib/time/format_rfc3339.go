package time

import stdtime "time"

// FormatRFC3339Millis formats t in ISO-8601 UTC with millisecond precision.
func FormatRFC3339Millis(t stdtime.Time) string { return t.UTC().Format(rfc3339Millis) }
