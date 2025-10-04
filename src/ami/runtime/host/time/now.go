package amitime

import stdtime "time"

// Now returns the current wall-clock time.
func Now() Time { return Time{t: stdtime.Now()} }

