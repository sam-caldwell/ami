package time

import stdtime "time"

// ParseRFC3339Millis parses an ISO-8601 UTC timestamp with millisecond precision.
func ParseRFC3339Millis(s string) (stdtime.Time, error) { return stdtime.Parse(rfc3339Millis, s) }
