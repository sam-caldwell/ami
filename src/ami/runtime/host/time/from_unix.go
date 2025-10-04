package amitime

import stdtime "time"

// FromUnix constructs a Time from seconds and nanoseconds since Unix epoch.
func FromUnix(sec, nsec int64) Time { return Time{t: stdtime.Unix(sec, nsec)} }

