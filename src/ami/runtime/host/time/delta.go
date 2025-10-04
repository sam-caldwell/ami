package amitime

// Delta returns t2 - t1 as a Duration. Positive when t2 occurs after t1.
func Delta(t1, t2 Time) Duration { return t2.t.Sub(t1.t) }

