package amitime

import stdtime "time"

// Duration is an alias to Go's time.Duration for convenience.
type Duration = stdtime.Duration

// Time wraps stdlib time.Time to provide AMI-friendly APIs.
type Time struct{ t stdtime.Time }

// Now returns the current wall-clock time.
func Now() Time { return Time{t: stdtime.Now()} }

// Sleep pauses the current goroutine for at least duration d.
func Sleep(d Duration) { stdtime.Sleep(d) }

// Add returns a new Time advanced by duration d from t.
func Add(t Time, d Duration) Time { return Time{t: t.t.Add(d)} }

// Delta returns t2 - t1 as a Duration. Positive when t2 occurs after t1.
func Delta(t1, t2 Time) Duration { return t2.t.Sub(t1.t) }

// FromUnix constructs a Time from seconds and nanoseconds since Unix epoch.
func FromUnix(sec, nsec int64) Time { return Time{t: stdtime.Unix(sec, nsec)} }

// Unix returns seconds since Unix epoch.
func (t Time) Unix() int64 { return t.t.Unix() }

// UnixNano returns nanoseconds since Unix epoch.
func (t Time) UnixNano() int64 { return t.t.UnixNano() }

