package time

import stdtime "time"

// Clock provides an injectable time source.
type Clock interface{ Now() stdtime.Time }

// FixedClock returns a constant instant.
type FixedClock struct{ T stdtime.Time }

// Now implements Clock.
func (c FixedClock) Now() stdtime.Time { return c.T }

