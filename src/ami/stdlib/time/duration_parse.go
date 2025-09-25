package time

import stdtime "time"

// ParseDuration parses a duration string, following Go's time.ParseDuration semantics.
func ParseDuration(s string) (stdtime.Duration, error) { return stdtime.ParseDuration(s) }

