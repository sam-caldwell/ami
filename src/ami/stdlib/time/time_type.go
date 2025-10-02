package amitime

import stdtime "time"

// Time wraps stdlib time.Time to provide AMI-friendly APIs.
type Time struct{ t stdtime.Time }

// Unix returns seconds since Unix epoch.
func (t Time) Unix() int64 { return t.t.Unix() }

// UnixNano returns nanoseconds since Unix epoch.
func (t Time) UnixNano() int64 { return t.t.UnixNano() }

